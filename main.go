package main

import (
    "os"
    "log"
    "strconv"
    "net/url"
    "net/http"
    "html/template"

    "gopkg.in/oauth2.v3"
    //"gopkg.in/oauth2.v3/errors"
    "gopkg.in/oauth2.v3/manage"
    "gopkg.in/oauth2.v3/models"
    "gopkg.in/oauth2.v3/server"
    "gopkg.in/oauth2.v3/store"
    "gopkg.in/session.v1"

    "github.com/go-oauth2/mongo"
    "github.com/go-oauth2/redis"
)

var (
    gSessions *session.Manager

    gClientID string = "76dbb7ac-6da8-11e6-84c6-1b976b623e41"
    gSecret string = "R7jjT7pwK3dhfjzrqhzRTmVXPJpmzwxqWFHg74bNgVdjnxg4d4FXCFxssFvTTgtt"
    gDomain string = "http://example.org:6379"
)

func init() {
    gSessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":3600}`)
    go gSessions.GC()
}


func main() {
    manager := manage.NewDefaultManager()
    // token memory store
    //manager.MustTokenStorage(store.NewMemoryTokenStore())
    // client test store
    client := &models.Client{
        ID:     gClientID,
        Secret: gSecret,
        Domain: gDomain,
    }
    manager.MapClientStorage(store.NewTestClientStore(client))

    // default redis store
    var (
        engine string = "redis"
        host string = "127.0.0.1"
        port int = 6379
        err error = nil
    )

    var (
        storage oauth2.TokenStore
        address string = host + ":" + strconv.Itoa(port)
    )
    switch engine {
    case "mongo":
        storage, err = mongo.NewTokenStore(mongo.NewConfig(
            "mongodb://" + address,
            "oauth2",
        ))
    case "redis":
        storage, err = redis.NewTokenStore(&redis.Config{
            Addr: address,
        })
    default:
        log.Println("Unsupported storage engine: ", engine)
        return
    }

    if err != nil {
        log.Println("NewTokenStore Error:", err.Error())
        return
    }
    manager.MustTokenStorage(storage, err)


    // new default server
    srv := server.NewDefaultServer(manager)
    srv.SetAllowGetAccessRequest(true)
    srv.SetInternalErrorHandler(func(err error) {
        log.Println("OAuth2 Error: ", err.Error())
    })

    // set hook handler
    srv.SetClientInfoHandler(ClientInfoHandler)
    srv.SetClientAuthorizedHandler(ClientAuthorizedHandler)
    srv.SetClientScopeHandler(ClientScopeHandler)
    srv.SetUserAuthorizationHandler(UserAuthorizationHandler)
    srv.SetPasswordAuthorizationHandler(PasswordAuthorizationHandler)

    // add static(css/img/js) handler
    http.Handle("/css/", http.FileServer(http.Dir("static")))
    http.Handle("/img/", http.FileServer(http.Dir("static")))
    http.Handle("/js/", http.FileServer(http.Dir("static")))

    // add http handler
    http.HandleFunc("/signup", SignupHandler)
    http.HandleFunc("/signin", SigninHandler)
    http.HandleFunc("/signout", SignoutHandler)
    http.HandleFunc("/auth", AuthHandler)
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
    var srvAddress string = ":6543"
    log.Println("Server is running at: ", srvAddress)
    log.Fatal(http.ListenAndServe(srvAddress, nil))
}


func SignupHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
    }
    outputHTML(w, r, "template/signup.html")
    return
}
func SigninHandler(w http.ResponseWriter, r *http.Request) {
    return
}
func SignoutHandler(w http.ResponseWriter, r *http.Request) {
    return
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        us, err := gSessions.SessionStart(w, r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        us.Set("UserID", "000000")
        w.Header().Set("Location", "/auth")
        w.WriteHeader(http.StatusFound)
        return
    }
    HtmlHandler(w, "template/signin.html")
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
    us, err := gSessions.SessionStart(w, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if us.Get("UserID") == nil {
        w.Header().Set("Location", "/login")
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

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
    file, err := os.Open(filename)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    defer file.Close()
    fi, _ := file.Stat()
    http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
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

