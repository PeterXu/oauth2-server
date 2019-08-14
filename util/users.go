package util

import (
	//"fmt"
	"log"
	"regexp"
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

func (u *Users) GetUserID(name string) (uid string, err error) {
	stmt := "SELECT uid FROM users WHERE username=? or email=? or cell=?"
	if err = u.db.QueryRow(stmt, name, name, name).Scan(&uid); err != nil {
		log.Printf("fail to get userid of (%s) - %s", name, err.Error())
		return
	}
	return
}

func (u *Users) CheckUser(name string) bool {
	var uid string
	stmt := "SELECT uid FROM users WHERE username=? or email=? or cell=?"
	err := u.db.QueryRow(stmt, name, name, name).Scan(&uid)
	if err == sql.ErrNoRows {
		//log.Printf("CheckUser - no username: %s in db", name)
		return false
	} else if err != nil {
		log.Printf("CheckUser - db error: %s", err.Error())
		return true
	}

	return true
}

// This password should be also hashed by md5 or sha in sender
func (u *Users) UpdatePassword(name, password string) (err error) {
	if !u.CheckUser(name) {
		err = ErrUserNotExist
		return
	}

	var salt_hash string
	if salt_hash, err = genPasswordHash(password); err != nil {
		return
	}

	stmt := "UPDATE users set password=? where username=? or email=? or cell=?"
	res, err := u.db.Exec(stmt, salt_hash, name, name, name)
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

func (u *Users) CreateUser(name, password string) (err error) {
	if u.CheckUser(name) {
		err = ErrUserExist
		return
	}

	var salt_hash string
	if salt_hash, err = genPasswordHash(password); err != nil {
		log.Println("CreateUser, genPasswordHash err=", err)
		return err
	}

	kind := getUserKind(name)
	stmt := "INSERT INTO users(uid, " + kind + ", password) VALUES (uuid(), ?, ?)"

	res, err := u.db.Exec(stmt, name, salt_hash)
	if err != nil {
		log.Println("CreateUser, db insert err=", err)
		return err
	}

	num, err := res.RowsAffected()
	if num != 1 || err != nil {
		err = ErrUserCreate
		return
	}
	return nil
}

// This password should be also hashed by md5 or sha in sender
func (u *Users) VerifyPassword(name, password string) (userID string, err error) {
	var db_salt_hash string

	stmt := "SELECT uid, password FROM users WHERE username=? or email=? or cell=?"
	err = u.db.QueryRow(stmt, name, name, name).Scan(&userID, &db_salt_hash)
	if err != nil {
		log.Printf("[VerifyPassword] fail to query name(%s): ", name, err.Error())
		return
	}

	if !checkPasswordHash(db_salt_hash, password) {
		err = errors.ErrAccessDenied
		log.Printf("[VerifyPassword] invalid password for name: %s", name)
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

func getUserKind(name string) string {
	if IsEmail(name) {
		return "email"
	}
	if IsCell(name) {
		return "cell"
	}
	return "username"
}

func IsEmail(email string) bool {
	if len(email) < 5 {
		return false
	}

	isOk, _ := regexp.MatchString("^[_a-z0-9-]+(\\.[_a-z0-9-]+)*@[a-z0-9-]+(\\.[a-z0-9-]+)*(\\.[a-z]{2,4})$", email)
	return isOk
}

func IsCell(cell string) bool {
	if len(cell) < 11 {
		return false
	}

	isOk, _ := regexp.MatchString(`^[\d]{11}$`, cell)
	return isOk
}
