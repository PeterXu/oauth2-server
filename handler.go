package main

import (
    "os"
    //"log"
    "bufio"
    "net/http"

    "gopkg.in/oauth2.v3"
    "gopkg.in/oauth2.v3/errors"
    "./util"
)

/// config in file(gClientId, gSecret)
func ClientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
    clientID = gClientID
    clientSecret = gSecret
    if clientID == "" || clientSecret == "" {
        err = errors.ErrInvalidClient
    }
    return
}

/// config in file
func ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
    err = nil
    allowed = true
    return
}

/// config in file
func ClientScopeHandler(clientID, scope string) (allowed bool, err error) {
    allowed = true
    return
}

func UserAuthorizationHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
    us, err := gSessions.SessionStart(w, r)
    uid := us.Get("UserID")
    if uid == nil {
        if r.Form == nil {
            r.ParseForm()
        }
        us.Set("Form", r.Form)
        w.Header().Set("Location", "/login")
        w.WriteHeader(http.StatusFound)
        return
    }
    userID = uid.(string)
    us.Delete("UserID")
    return
}

func PasswordAuthorizationHandler(username, password string) (userID string, err error) {
    htpasswd, err := CheckPassword("./passwd")
    if err != nil {
        err = util.ErrInvalidPassword
        return
    }

    err = htpasswd.AuthenticateUser(username, password)
    userID = "123456" + username
    return
}

func CheckPassword(filename string) (htpasswd *util.HTPasswd, err error) {
    file, err := os.Open(filename)
    if err != nil{
        return
    }
    defer file.Close()
    rio := bufio.NewReader(file)
    htpasswd, err = util.NewHTPasswd(rio)
    return
}

