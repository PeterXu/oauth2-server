package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tabalt/gracehttp"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/session.v1"

	"github.com/PeterXu/oauth2-server/util"
)

type Global struct {
	Sessions *session.Manager
	Users    *util.Users
	Config   Config
	Server   *server.Server
	Store    *TokenStoreX
	Hub      *Hub
}

var gg Global

func init() {
	gg.Sessions, _ = session.NewManager("memory", `{"cookieName":"gosessionid","gclifetime":3600}`)
	go gg.Sessions.GC()
}

func main() {
	var err error

	/// read & parse config
	var fname string
	flag.StringVar(&fname, "c", kDefaultConfig, "server config file")
	flag.Parse()

	conf, err := NewConfig(fname)
	if err != nil {
		log.Fatal("[main] config err: ", err)
		return
	}
	//log.Println("[main] config: ", conf)

	/// Start service
	if conf.Service.Enable {
		gg.Hub = newHub()
		go gg.Hub.run()
		go NewService(conf.Service)
	}

	manager := manage.NewDefaultManager()
	// token memory store
	//manager.MustTokenStorage(store.NewMemoryTokenStore())
	manager.MapClientStorage(NewMyClientStore(conf.Clients))

	/// default redis store
	store, err := NewTokenStore(conf.Store)
	if err != nil {
		log.Println("[main] fail to NewTokenStore: ", err.Error())
		return
	}
	manager.MustTokenStorage(store, err)

	/// init users DB
	users := util.NewUsers(conf.Db.Engine, conf.Db.Connection)
	if users == nil {
		log.Fatal("[main] fail to NewUsers (init db)")
		return
	}

	/// new default server
	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetInternalErrorHandler(InternalErrorHandler)

	/// set internel hook handler
	srv.SetClientInfoHandler(ClientInfoHandler)
	srv.SetClientAuthorizedHandler(ClientAuthorizedHandler)
	srv.SetClientScopeHandler(ClientScopeHandler)
	srv.SetUserAuthorizationHandler(UserAuthorizationHandler)
	srv.SetPasswordAuthorizationHandler(PasswordAuthorizationHandler)
	srv.SetAccessTokenExpHandler(AccessTokenExpHandler)

	/// add static(css/img/js) handler
	http.Handle("/css/", http.FileServer(http.Dir("static")))
	http.Handle("/img/", http.FileServer(http.Dir("static")))
	http.Handle("/js/", http.FileServer(http.Dir("static")))

	/// add http handler
	http.HandleFunc("/reset", ResetHandler)
	http.HandleFunc("/signup", SignupHandler)
	http.HandleFunc("/signin", SigninHandler)
	http.HandleFunc("/signout", SignoutHandler)
	http.HandleFunc("/auth", AuthHandler)
	http.HandleFunc("/code", CodeHandler)
	http.HandleFunc("/check", CheckHandler)
	http.HandleFunc("/api/", ApiHandler)
	http.HandleFunc("/", NotFoundHandler)

	// called by HandleAuthorizeRequest
	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[main], authorize begin")
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	/// called by HandleTokenRequest
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[main], token begin")
		err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	/// save global variables
	gg.Config = conf
	gg.Users = users
	gg.Server = srv
	gg.Store = store

	/// start http server
	address := conf.Listen.Host + ":" + strconv.Itoa(conf.Listen.Port)
	log.Println("[main] Server is running at: ", address)
	log.Fatal(gracehttp.ListenAndServe(address, nil))
}

func ResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.FormValue("username"))
		password := strings.TrimSpace(r.FormValue("password"))
		if len(username) < kMinUsernameLength || len(password) < kMinPasswordLength {
			ResponseErrorWithJson(w, errors.ErrInvalidRequest)
			return
		}

		password1 := strings.TrimSpace(r.FormValue("password1"))
		password2 := strings.TrimSpace(r.FormValue("password2"))
		if len(password1) < kMinPasswordLength || password1 != password2 {
			ResponseErrorWithJson(w, errors.ErrInvalidRequest)
			return
		}

		_, err := gg.Users.VerifyPassword(username, password)
		if err != nil {
			ResponseErrorWithJson(w, errors.ErrAccessDenied)
			return
		}

		err = gg.Users.UpdatePassword(username, password1)
		if err != nil {
			ResponseErrorWithJson(w, errors.ErrServerError)
			return
		}

		return
	}

	HtmlHandler(w, "template/reset.html")
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.FormValue("username"))
		if len(username) < kMinUsernameLength {
			ResponseErrorWithJson(w, errors.ErrInvalidRequest)
			return
		}

		password1 := strings.TrimSpace(r.FormValue("password1"))
		password2 := strings.TrimSpace(r.FormValue("password2"))
		if len(password1) < kMinPasswordLength || password1 != password2 {
			ResponseErrorWithJson(w, errors.ErrInvalidRequest)
			return
		}

		err := gg.Users.CreateUser(username, password1)
		if err != nil {
			ResponseErrorWithJson(w, errors.ErrServerError)
			return
		}
		Response200SuccessWithJson(w)
		return
	}
	HtmlHandler(w, "template/signup.html")
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		us, err := gg.Sessions.SessionStart(w, r)
		if err != nil {
			log.Printf("SigninHandler, err=%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		password := strings.TrimSpace(r.FormValue("password"))
		if len(username) < kMinUsernameLength || len(password) < kMinPasswordLength {
			ResponseErrorWithJson(w, errors.ErrInvalidRequest)
			return
		}

		uid, err := gg.Users.VerifyPassword(username, password)
		if err != nil {
			ResponseErrorWithJson(w, errors.ErrAccessDenied)
			return
		}

		us.Set("UserID", uid)

		// required: client_id/response_type/state/scope
		// optional: redirect_uri,
		if len(r.FormValue("client_id")) <= 0 {
			r.Form.Set("client_id", kDefaultClientID)
		}
		if len(r.FormValue("response_type")) <= 0 {
			// default use "token" not "code".
			r.Form.Set("response_type", oauth2.Token.String())
		}

		us.Set("Form", r.Form)

		// (a) for standard flow: user-allow required, http GET(/auth? -> /authorize?)
		// (b) non-standard flow: jump to authorize directly, http POST
		if gg.Config.Flow == "direct" {
			u := new(url.URL)
			u.Path = "/authorize"
			u.RawQuery = r.Form.Encode()
			w.Header().Set("Location", u.String())
			w.WriteHeader(http.StatusFound)
		} else {
			u := new(url.URL)
			u.Path = "/auth"
			u.RawQuery = r.Form.Encode()
			w.Header().Set("Location", u.String())
			w.WriteHeader(http.StatusFound)
		}
		return
	}
	HtmlHandler(w, "template/signin.html")
}

func SignoutHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: signout and disable current token
	HtmlHandler(w, "template/signout.html")
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	us, err := gg.Sessions.SessionStart(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if us.Get("UserID") == nil {
		w.Header().Set("Location", "/signin")
		w.WriteHeader(http.StatusFound) // 302
		return
	}

	if r.Method == http.MethodPost {
		if us.Get("Form") == nil {
			http.Error(w, util.ErrInvalidRequestArgs.Error(), http.StatusBadRequest)
			return
		}

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
	if r.Method == http.MethodPost {
		var rtype oauth2.ResponseType
		response_type := strings.TrimSpace(r.FormValue("response_type"))
		if len(response_type) == 0 || response_type == "code" {
			rtype = oauth2.Code
		} else if response_type == "token" {
			rtype = oauth2.Token
		} else {
			ResponseErrorWithJson(w, errors.ErrInvalidRequest)
			return
		}

		var uid string
		username := strings.TrimSpace(r.FormValue("username"))
		if len(username) == 0 {
			uid = gg.Store.randUid()
			//log.Println("new temporay uid=", uid)
		} else {
			password := strings.TrimSpace(r.FormValue("password"))
			if len(username) < kMinUsernameLength || len(password) < kMinPasswordLength {
				ResponseErrorWithJson(w, errors.ErrInvalidRequest)
				return
			}

			var err error
			if uid, err = gg.Users.VerifyPassword(username, password); err != nil {
				ResponseErrorWithJson(w, errors.ErrAccessDenied)
				return
			}
		}

		exp, err := strconv.Atoi(r.FormValue("exp"))
		if err != nil {
			exp = 0
		}

		clientID := strings.TrimSpace(r.FormValue("client_id"))
		if len(clientID) <= 0 {
			clientID = kDefaultClientID
		}

		cli, err := gg.Server.Manager.GetClient(clientID)
		if err != nil {
			ResponseErrorWithJson(w, errors.ErrInvalidClient)
			return
		}
		redirectURI := cli.GetDomain()
		//log.Println(redirectURI)

		req := &server.AuthorizeRequest{
			UserID:         uid,
			RedirectURI:    redirectURI,
			ResponseType:   rtype,
			ClientID:       clientID,
			State:          r.FormValue("state"),
			Scope:          r.FormValue("scope"),
			AccessTokenExp: time.Second * time.Duration(exp),
		}

		tgr := &oauth2.TokenGenerateRequest{
			ClientID:       req.ClientID,
			UserID:         req.UserID,
			RedirectURI:    req.RedirectURI,
			Scope:          req.Scope,
			AccessTokenExp: req.AccessTokenExp,
		}

		ti, err := gg.Server.Manager.GenerateAuthToken(req.ResponseType, tgr)
		if err != nil {
			ResponseErrorWithJson(w, errors.ErrServerError)
			return
		}

		data := make(map[string]interface{})
		if rtype == oauth2.Code {
			data["code"] = ti.GetCode()
			data["expires_in"] = int64(ti.GetCodeExpiresIn() / time.Second)
		} else {
			data["access_token"] = ti.GetAccess()
			data["expires_in"] = int64(ti.GetAccessExpiresIn() / time.Second)
			data["refresh_token"] = ti.GetRefresh()
		}

		if req.State != "" {
			data["state"] = req.State
		}
		if req.Scope != "" {
			data["scope"] = req.Scope
		}

		//log.Printf("[CodeHandler] data=", data)
		ResponseDataWithJson(w, data, http.StatusOK)
	}
}

func CheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Support no username
		username := strings.TrimSpace(r.FormValue("username"))
		if len(username) > 0 {
			if len(username) < kMinUsernameLength {
				ResponseErrorWithJson(w, errors.ErrInvalidRequest)
				return
			}
		}

		scope := strings.TrimSpace(r.FormValue("scope"))

		// Should have one of: access_token or refresh_token
		var access_token, refresh_token string
		access_token = strings.TrimSpace(r.FormValue("token"))
		if len(access_token) == 0 {
			access_token = strings.TrimSpace(r.FormValue("access_token"))
			if len(access_token) == 0 {
				refresh_token = strings.TrimSpace(r.FormValue("refresh_token"))
				if len(refresh_token) == 0 {
					ResponseErrorWithJson(w, errors.ErrInvalidRequest)
					return
				}
			}
		}

		var err error
		var ti oauth2.TokenInfo
		if len(access_token) > 0 {
			ti, err = gg.Server.Manager.LoadAccessToken(access_token)
		} else {
			ti, err = gg.Server.Manager.LoadRefreshToken(refresh_token)
		}

		if err != nil {
			ResponseErrorWithJson(w, err)
			return
		}

		if scope != ti.GetScope() {
			ResponseErrorWithJson(w, errors.ErrInvalidRequest)
			return
		}

		uid := ti.GetUserID()
		if len(username) == 0 {
			// check uid in Store
			//log.Println("uid:", uid)
			if err = gg.Store.checkUid(uid); err != nil {
				ResponseErrorWithJson(w, errors.ErrInvalidRequest)
				return
			}
		} else {
			// check uid in Users
			if !gg.Users.CheckUserID(uid, username) {
				ResponseErrorWithJson(w, errors.ErrInvalidRequest)
				return
			}
		}
		Response200SuccessWithJson(w)
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
		log.Println("[HtmlHandler] error: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, nil)
}

func ResponseDataWithJson(w http.ResponseWriter, data map[string]interface{}, status int) (err error) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(data)
	return
}

func ResponseErrorWithJson(w http.ResponseWriter, respErr error) (err error) {
	data, status, _ := gg.Server.GetErrorData(respErr)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(data)
	return
}

func Response200DataWithJson(w http.ResponseWriter, data map[string]interface{}) (err error) {
	return ResponseDataWithJson(w, data, 200)
}

func Response200SuccessWithJson(w http.ResponseWriter) (err error) {
	return Response200DataWithJson(w, map[string]interface{}{
		"status": "Success",
	})
}

func Response400ErrorWithJson(w http.ResponseWriter, respErr error) (err error) {
	data := map[string]interface{}{
		"status": respErr.Error(),
	}
	return ResponseDataWithJson(w, data, 400)
}
