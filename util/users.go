package util


import (
    "fmt"
    "log"
    "strconv"
    "strings"
    "database/sql"

    "gopkg.in/oauth2.v3/errors"
    _ "github.com/go-sql-driver/mysql"
)


type Users struct {
    db *sql.DB
    dbtype string
    dbconn string
}

func NewUsers(dbtype string, dbconn string) *Users {
    users := &Users{
        db: nil,
        dbtype: dbtype,
        dbconn: dbconn,
    }

    db := initDB(dbtype, dbconn)
    if db == nil {
        return nil
    }

    users.db = db
    return users
}

func (u *Users) GetUserID(username string) (uid string, err error){
    stmt := "SELECT uid FROM users WHERE username = ?"
    err = u.db.QueryRow(stmt, username).Scan(&uid)
    if err != nil {
        log.Printf("fail to get userid of (%s) - %s", username, err.Error())
        return
    }

    return
}

func (u *Users) CheckUser(username string) bool {
    var uid string
    stmt := "SELECT uid FROM users WHERE username=?"
    err := u.db.QueryRow(stmt, username).Scan(&uid)
    if err == sql.ErrNoRows {
        log.Printf("[CheckUser] - no username=%s in db", username)
        return false
    }else if err != nil {
        log.Fatal("[CheckUser] - db error=", err.Error())
        return true
    }

    return true
}

func (u *Users) CreateUser(username, password string) (err error){
    if u.CheckUser(username) {
        err = ErrUserExist
        return
    }

    cnt := kDefaultIter
    salt, hash := hashPasswordNoSalt(password, cnt)
    if len(salt) <= 0 || len(hash) <= 0 {
        err = ErrFailed
        return
    }

    salt_hash := fmt.Sprintf("self:sha256:%d$%s$%s", cnt, salt, hash)
    stmt := "INSERT INTO users(uid, username, password) VALUES (uuid(), ?, ?)"
    res, err := u.db.Exec(stmt, username, salt_hash)
    if err != nil {
        return
    }

    num, err := res.RowsAffected()
    if num != 1 || err != nil {
        err = ErrUserCreate
        return
    }

    return
}

// This password should be also hashed by md5 or sha in sender
func (u *Users) VerifyPassword(username, password string) (err error){
    var db_salt_hash string
    stmt := "SELECT password FROM users WHERE username = ?"
    err = u.db.QueryRow(stmt, username).Scan(&db_salt_hash)
    if err != nil {
        log.Println("fail to query sql: ", err.Error())
        return
    }

    str := strings.Split(db_salt_hash, "$")
    if len(str) != 3 {
        err = ErrServerFault
        log.Println("invalid password(salt+hash) from db: ", db_salt_hash)
        return
    }

    db_cnt := 1
    str2 := strings.Split(str[0], ":")
    if len(str2) == 3 {
        db_cnt, _ = strconv.Atoi(str2[2])
    }

    db_salt := str[1]
    db_password := str[2]

    hash := hashPasswordWithSalt(db_salt, password, db_cnt)
    if hash != db_password {
        err = errors.ErrAccessDenied
        return
    }

    return
}

func (u *Users) Close() {
    if u.db != nil {
        u.db.Close()
    }
}


func initDB(engine, conn string) *sql.DB{
    db, err := sql.Open(engine, conn)
    if err != nil {
        log.Println("fail to open db: %s - %s ", conn, err.Error())
        return nil
    }
    //defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Println("fail to ping db: ", err.Error())
        return nil
    }

    return db
}

