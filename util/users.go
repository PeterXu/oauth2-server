package util


import (
    "log"
    "bytes"
    //"crypto/md5"
    "crypto/hmac"
    "crypto/sha256"
    "database/sql"

    "gopkg.in/oauth2.v3/errors"
    _ "github.com/go-sql-driver/mysql"
)

var (
    gSaltKey []byte = []byte("h7FkjHcgWTxKqVtVqj9WhLcxgz3rw3zvrmtFfsdKcCKxnTKPdb99hJjnNgxN9std")
    gDbEngine string = "mysql"
    //gDbConn string = "oauth:oauth@tcp(127.0.0.1:3306)/oauth"
    gDbConn string = "oauth:oauth@/oauth"

    gDB *sql.DB = nil
)

type User struct {
    uid string
    username string
}

func init() {
    db, err := sql.Open(gDbEngine, gDbConn)
    if err != nil {
        log.Println("fail to open db: %s - %s ", gDbConn, err.Error())
        return
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Println("fail to ping db: ", err.Error())
        return
    }

    gDB = db
    return
}

func GetUserID(uname string) (uid string, err error){
    if gDB == nil {
        log.Println("db is not opened: %s ", gDbConn)
        return
    }

    stmt, err := gDB.Prepare("select uid from users where username = ?")
    if err != nil {
        log.Println("fail to prepare sql: ", err.Error())
        return
    }
    defer stmt.Close()

    err = stmt.QueryRow(uname).Scan(&uid)
    if err != nil {
        log.Println("fail to query sql: ", err.Error())
        return
    }

    return
}

// This password should be also hashed by md5 or sha in sender
func VerifyPassword(username, password string) (err error){
    if gDB == nil {
        log.Println("db is not opened: %s ", gDbConn)
        return
    }

    stmt, err := gDB.Prepare("select password from users where username = ?")
    if err != nil {
        log.Println("fail to prepare sql: ", err.Error())
        return
    }
    defer stmt.Close()

    var db_password string
    err = stmt.QueryRow(username).Scan(&db_password)
    if err != nil {
        log.Println("fail to query sql: ", err.Error())
        return
    }

    hash_password, _ := hashPassword(username, password)
    if hash_password != db_password {
        err = errors.ErrAccessDenied
        return
    }

    return
}

func genSalt(message, key []byte) (salt []byte) {
    mac := hmac.New(sha256.New, key)
    mac.Write(message)
    salt = mac.Sum(nil)[:8]
    return
}

func hashPassword(username, password string) (hash, salt string) {
    message := []byte(username)
    salt = string(genSalt(message, gSaltKey))   // 8bytes
    sha := sha256.Sum256([]byte(salt+password)) // 32bytes
    hash = bytes.NewBuffer(sha[:]).String()
    return
}


