package util

import (
	//"fmt"
	"log"
	//"strconv"
	//"strings"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/oauth2.v3/errors"
)

type Users struct {
	db     *sql.DB
	dbtype string
	dbconn string
}

//
// @dbtype: 'mysql' only now
// @dbconn: the format,  "user:pass@[proto(addr)]/database",
//      (a) "oauth:oauth@tcp(127.0.0.1:3306)/oauth",
//      (b) "oauth:oauth@/oauth",
//      (c) "oauth:oauth@unix(/var/run/mysql.sock)/oauth",
func NewUsers(dbtype string, dbconn string) *Users {
	var db *sql.DB
	if db = initDB(dbtype, dbconn); db == nil {
		return nil
	}

	return &Users{
		db:     db,
		dbtype: dbtype,
		dbconn: dbconn,
	}
}

func (u *Users) GetUserID(username string) (uid string, err error) {
	stmt := "SELECT uid FROM users WHERE username = ?"
	if err = u.db.QueryRow(stmt, username).Scan(&uid); err != nil {
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
		log.Printf("[CheckUser] - no username: %s in db", username)
		return false
	} else if err != nil {
		log.Printf("[CheckUser] - db error: %s", err.Error())
		return true
	}

	return true
}

// This password should be also hashed by md5 or sha in sender
func (u *Users) UpdatePassword(username, password string) (err error) {
	if !u.CheckUser(username) {
		err = ErrUserNotExist
		return
	}

	var salt_hash string
	if salt_hash, err = genPasswordHash(password); err != nil {
		return
	}

	stmt := "UPDATE users set password=? where username=?"
	res, err := u.db.Exec(stmt, salt_hash, username)
	if err != nil {
		return
	}

	num, err := res.RowsAffected()
	if num != 1 || err != nil {
		err = ErrUpdatePassword
		return
	}

	return
}

func (u *Users) CreateUser(username, password string) (err error) {
	if u.CheckUser(username) {
		err = ErrUserExist
		return
	}

	var salt_hash string
	if salt_hash, err = genPasswordHash(password); err != nil {
		return
	}

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
func (u *Users) VerifyPassword(username, password string) (userID string, err error) {
	var db_salt_hash string
	stmt := "SELECT uid, password FROM users WHERE username = ?"
	err = u.db.QueryRow(stmt, username).Scan(&userID, &db_salt_hash)
	if err != nil {
		log.Println("[VerifyPassword] fail to query sql: ", err.Error())
		return
	}

	if !checkPasswordHash(db_salt_hash, password) {
		err = errors.ErrAccessDenied
		log.Printf("[VerifyPassword] invalid password for username: %s", username)
	}

	return
}

func (u *Users) Close() {
	if u.db != nil {
		u.db.Close()
		u.db = nil
	}
}

func initDB(engine, conn string) *sql.DB {
	db, err := sql.Open(engine, conn)
	if err != nil {
		log.Println("[initDB] fail to open db: %s - %s ", conn, err.Error())
		return nil
	}
	//defer db.Close()

	if err = db.Ping(); err != nil {
		log.Println("[initDB] fail to ping db: ", err.Error())
		return nil
	}

	return db
}
