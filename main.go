package main

import (
    "log"
    "strconv"
    "net/http"

    "github.com/go-oauth2/mongo"
    "github.com/go-oauth2/redis"

    "gopkg.in/oauth2.v3"
    "gopkg.in/oauth2.v3/manage"
    "gopkg.in/oauth2.v3/server"
    "gopkg.in/oauth2.v3/store"
)

func main() {
    manager := manage.NewDefaultManager()
    // token memory store
    manager.MustTokenStorage(store.NewMemoryTokenStore())
    // client test store
    manager.MapClientStorage(store.NewTestClientStore())

    // default redis store
    var (
        engine string = "redis"
        host string = "127.0.0.1"
        port int = 6379
    )

    var (
        err error
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


    //> new default server
    srv := server.NewDefaultServer(manager)
    srv.SetAllowGetAccessRequest(true)
    srv.SetInternalErrorHandler(func(err error) {
        log.Println("OAuth2 Error: ", err.Error())
    })

    //> http handler
    http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
        err := srv.HandleAuthorizeRequest(w, r)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
    })

    http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
        srv.HandleTokenRequest(w, r)
    })

    //> start http server
    http.ListenAndServe(":6543", nil)
}

