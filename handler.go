package main

import (
    "os"
    "log"
    "bufio"
    "net/http"

    "gopkg.in/oauth2.v3"
    "gopkg.in/oauth2.v3/errors"
    "./util"
)

/// config in file(gClientId, gSecret)
func ClientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
    log.Println("[ClientInfoHandler] ..")
    clientID = gClientID
    clientSecret = gSecret
    if clientID == "" || clientSecret == "" {
        err = errors.ErrInvalidClient
    }
    return
}

/// config in file
func ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
    log.Println("[ClientAuthorizedHandler] ..")
    err = nil
    allowed = true
    return
}

/// config in file
func ClientScopeHandler(clientID, scope string) (allowed bool, err error) {
    log.Println("[ClientScopeHandler] ..")
    allowed = true
    return
}

func UserAuthorizationHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
    log.Println("[UserAuthorizationHandler] ..")
    return
}

func PasswordAuthorizationHandler(username, password string) (userID string, err error) {
    log.Println("[PasswordAuthorizationHandler] ..")

    // htpasswd, err := CheckPassword("./passwd")
    // err = htpasswd.AuthenticateUser(username, password)
    err = gUsers.VerifyPassword(username, password)
    if err == nil {
        userID, err = gUsers.GetUserID(username)
    }

    if err != nil {
        log.Printf("[PasswordAuthorizationHandler] userID=%s, err=%s", userID, err)
    }

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

