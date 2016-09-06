package main

import (
    "log"
    "time"
    "strconv"
    "strings"
    "net/url"
    "net/http"
    "html/template"

    "gopkg.in/oauth2.v3"
    "gopkg.in/oauth2.v3/manage"
    "gopkg.in/oauth2.v3/server"
    "gopkg.in/session.v1"

    "./util"
)

var (
    gSessions *session.Manager
    gUsers *util.Users
    gConfig Config
    gServer *server.Server
)

func init() {
    gSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":3600}`)
    go gSessions.GC()
}

func main() {
    var err error

    fname := "./server.toml"
    conf, err := NewConfig(fname)
    if err != nil {
        log.Fatal("[main] config err: ", err)
        return
    }
    gConfig = conf
    //log.Println("[main] config: ", conf)


    manager := manage.NewDefaultManager()
    // token memory store
    //manager.MustTokenStorage(store.NewMemoryTokenStore())
    manager.MapClientStorage(NewMyClientStore(conf.Clients))

    // default redis store
    storage, err := NewTokenStore(conf.Store)
    if err != nil {
        log.Println("[main] NewTokenStore Error:", err.Error())
        return
    }
    manager.MustTokenStorage(storage, err)

    // init users DB
    gUsers = util.NewUsers(conf.Db.Engine, conf.Db.Connection)
    if gUsers == nil {
        return
    }

    // new default server
    srv := server.NewDefaultServer(manager)
    gServer = srv

    srv.SetAllowGetAccessRequest(true)
    srv.SetInternalErrorHandler(func(err error) {
        log.Println("[main] OAuth2 Error: ", err.Error())
    })

    // set hook handler
    srv.SetClientInfoHandler(ClientInfoHandler)
    srv.SetClientAuthorizedHandler(ClientAuthorizedHandler)
    srv.SetClientScopeHandler(ClientScopeHandler)
    srv.SetUserAuthorizationHandler(UserAuthorizationHandler)
    srv.SetPasswordAuthorizationHandler(PasswordAuthorizationHandler)
    srv.SetAccessTokenExpHandler(AccessTokenExpHandler)

    // add static(css/img/js) handler
    http.Handle("/css/", http.FileServer(http.Dir("static")))
    http.Handle("/img/", http.FileServer(http.Dir("static")))
    http.Handle("/js/", http.FileServer(http.Dir("static")))

    // add http handler
    http.HandleFunc("/signup", SignupHandler)
    http.HandleFunc("/signin", SigninHandler)
    http.HandleFunc("/signout", SignoutHandler)
    http.HandleFunc("/auth", AuthHandler)
    http.HandleFunc("/code", CodeHandler)
    http.HandleFunc("/", NotFoundHandler)

    // called by HandleAuthorizeRequest
    http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
        err := srv.HandleAuthorizeRequest(w, r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
    })

    // callbed by HandleTokenRequest
    http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
        err := srv.HandleTokenRequest(w, r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    })


    // start http server
    srvAddress := conf.Listen.Host + ":" + strconv.Itoa(conf.Listen.Port)
    log.Println("Server is running at: ", srvAddress)
    log.Fatal(http.ListenAndServe(srvAddress, nil))
}


func SignupHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        username := strings.TrimSpace(r.FormValue("username"))
        password := strings.TrimSpace(r.FormValue("password"))
        if len(username) < 6 || len(password) < 6 {
            log.Printf("[SignupHandler] invalid username/password: %s/%s", username, password)
            http.Error(w, "Invalid username or password", 400)
            return
        }

        err := gUsers.CreateUser(username, password)
        if err != nil {
            log.Printf("[SignupHandler] fail to create username=%s, err=%s", username, err.Error())
            http.Error(w, "Fail to create username: "+username, 400)
            return
        }
        return
    }
    HtmlHandler(w, "template/signup.html")
    return
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        us, err := gSessions.SessionStart(w, r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        username := strings.TrimSpace(r.FormValue("username"))
        password := strings.TrimSpace(r.FormValue("password"))
        if len(username) < 6 || len(password) < 6 {
            log.Printf("[SignupHandler] invalid username/password: %s/%s", username, password)
            http.Error(w, "Invalid username or password", 400)
            return
        }

        uid, err := gUsers.VerifyPassword(username, password)
        if err != nil {
            log.Printf("[SignupHandler] fail to verify username=%s, err=%s", username, err.Error())
            http.Error(w, "Fail to verify username: " + username, 403)
            return
        }

        us.Set("UserID", uid)
        w.Header().Set("Location", "/auth")
        w.WriteHeader(http.StatusFound)
        return
    }
    HtmlHandler(w, "template/signin.html")
}

func SignoutHandler(w http.ResponseWriter, r *http.Request) {
    HtmlHandler(w, "template/signout.html")
    return
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
    us, err := gSessions.SessionStart(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if us.Get("UserID") == nil {
        w.Header().Set("Location", "/signin")
        w.WriteHeader(http.StatusFound)
        return
    }

    if r.Method == "POST" {
        form := us.Get("Form").(url.Values)
        u := new(url.URL)
        u.Path = "/authorize"
        u.RawQuery = form.Encode()
        w.Header().Set("Location", u.String())
        w.WriteHeader(http.StatusFound)
        us.Delete("Form")
        return
    }
    HtmlHandler(w, "template/auth.html")
}

func CodeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        username := strings.TrimSpace(r.FormValue("username"))
        password := strings.TrimSpace(r.FormValue("password"))
        if len(username) < 6 || len(password) < 6 {
            log.Printf("[CodeHandler] invalid username/password: %s/%s", username, password)
            http.Error(w, "Invalid username or password", 400)
            return
        }

        uid, err := gUsers.VerifyPassword(username, password)
        if err != nil {
            log.Printf("[CodeHandler] fail to verify username=%s, err=%s", username, err.Error())
            http.Error(w, "Fail to verify username: " + username, 403)
            return
        }

        clientID := strings.TrimSpace(r.FormValue("client_id"))
        if len(clientID) <= 1 {
            clientID = kDefaultClientID
        }

        cli, err := gServer.Manager.GetClient(clientID)
        if err != nil {
            log.Printf("[CodeHandler] invalid client_id: ", clientID)
            http.Error(w, "Invalid client", 403)
            return
        }
        redirectURI := cli.GetDomain()

        req := &server.AuthorizeRequest{
            UserID:       uid,
            RedirectURI:  redirectURI,
            ResponseType: "code",
            ClientID:     clientID,
            State:        r.FormValue("state"),
            Scope:        r.FormValue("scope"),
            AccessTokenExp: time.Second * 3600,
        }

        tgr := &oauth2.TokenGenerateRequest{
            ClientID:       req.ClientID,
            UserID:         req.UserID,
            RedirectURI:    req.RedirectURI,
            Scope:          req.Scope,
            AccessTokenExp: req.AccessTokenExp,
        }

        ti, err := gServer.Manager.GenerateAuthToken(req.ResponseType, tgr)
        if err != nil {
            log.Printf("[CodeHandler] err=", err.Error())
            http.Error(w, "Fail to generate token: " + err.Error(), 403)
            return
        }
        log.Printf("[CodeHandler] code=", ti.GetCode())
    }
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path == "/" {
        http.Redirect(w, r, "/signin", http.StatusFound)
        return
    }

    HtmlHandler(w, "template/error/404.html")
}

func HtmlHandler(w http.ResponseWriter, filename string) {
    t, err := template.ParseFiles(filename)
    if err != nil {
        log.Println(err)
        http.Error(w, err.Error(), 500)
        return
    }

    t.Execute(w, nil)
}

