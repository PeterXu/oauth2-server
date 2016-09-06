package main

import (
    "os"
    "log"
    "time"
    "bufio"
    "strings"
    "net/http"

    "gopkg.in/oauth2.v3"
    "gopkg.in/oauth2.v3/errors"
    "./util"
)

/// config in file(ClientId, Secret)
func ClientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
    //log.Println("[ClientInfoHandler] req: ", r)
    clientID = strings.TrimSpace(r.FormValue("client_id"))

    if len(clientID) <= 1 {
        clientID = kDefaultClientID
        clientSecret = kDefaultClientSecret
        //log.Println("[ClientInfoHandler] ", err, clientSecret)
    }else {
        cinfo, err := gConfig.GetClientByID(clientID)
        if err == nil {
            clientSecret = cinfo.Secret
        }
    }

    if clientID == "" || clientSecret == "" {
        err = errors.ErrInvalidClient
    }
    return
}

/// config in file
func ClientAuthorizedHandler(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
    szGrant := grant.String()
    log.Printf("[ClientAuthorizedHandler] clientID: %s, grant: %s", clientID, szGrant)

    if len(szGrant) <= 0 {
        return
    }

    if clientID == kDefaultClientID {
        allowed = true
        return
    }

    cinfo, err := gConfig.GetClientByID(clientID)
    if err != nil {
        log.Printf("[ClientAuthorizedHandler] no info for client: %s", clientID)
        return
    }

    for _, val := range cinfo.Grants {
        if val == szGrant {
            allowed = true
            break
        }
    }
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
    us, err := gSessions.SessionStart(w, r)
    uid := us.Get("UserID")
    if uid == nil {
        if r.Form == nil {
            r.ParseForm()
        }
        us.Set("Form", r.Form)
        w.Header().Set("Location", "/signin")
        w.WriteHeader(http.StatusFound)
        return
    }

    userID = uid.(string)
    us.Delete("UserID")
    return
}

func PasswordAuthorizationHandler(username, password string) (userID string, err error) {
    log.Println("[PasswordAuthorizationHandler] ..")

    // htpasswd, err := CheckPassword("./passwd")
    // err = htpasswd.AuthenticateUser(username, password)
    userID, err = gUsers.VerifyPassword(username, password)
    if err != nil {
        log.Printf("[PasswordAuthorizationHandler] userID=%s, err=%s", userID, err.Error())
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

func AccessTokenExpHandler(w http.ResponseWriter, r *http.Request) (exp time.Duration, err error) {
    exp = time.Second * 3600
    return
}

