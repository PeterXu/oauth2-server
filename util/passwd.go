package util

import (
    "log"
    "fmt"
    "strings"
    "strconv"
    "errors"
    //"bytes"
    //"encoding/base64"
    crand "crypto/rand"
    //"crypto/md5"
    //"crypto/hmac"
    "crypto/sha1"
    "crypto/sha256"
    "crypto/sha512"
    "encoding/hex"

    "golang.org/x/crypto/pbkdf2"
)

const kDefaultIter int = 1000


func genRand(num int) []byte{
    b := make([]byte, num)
    _, err := crand.Read(b)
    if err != nil {
        return nil
    }
    return b
}

func genSalt(num int) string {
    return hex.EncodeToString(genRand(16))[0:num]
}

func pbkdf2_hash(name string, password []byte, salt []byte, iterations int) []byte {
    var rk []byte
    switch name {
    case "sha1":
        rk = pbkdf2.Key(password, salt, iterations, 20, sha1.New)
    case "sha256":
        rk = pbkdf2.Key(password, salt, iterations, 20, sha256.New)
    case "sha512":
        rk = pbkdf2.Key(password, salt, iterations, 20, sha512.New)
    }
    return rk
}

func pbkdf2_validate(password string, hash string) (bool, error) {
    parts := strings.Split(hash, "$")
    if len(parts) != 3 {
        log.Println("[pbkdf2_validate] invalid hash: ", hash)
        return false, errors.New("hash is not a pbkdf2")
    }

    head := strings.Split(parts[0], ":")
    if len(head) != 3 {
        log.Println("[pbkdf2_validate] invalid parts: ", parts[0])
        return false, errors.New("hash is not a pbkdf2")
    }

    if head[0] != "pbkdf2" {
        log.Println("[pbkdf2_validate] invalid mark: ", head[0])
        return false, errors.New("hash is not a pbkdf2")
    }

    //log.Printf("[pbkdf2_validate] hash: [%s][%s][%s]", head[1], parts[1], head[2])
    method := head[1]
    iterations, _ := strconv.Atoi(head[2])
    salt := []byte(parts[1])
    rk := pbkdf2_hash(method, []byte(password), salt, iterations)

    result := hex.EncodeToString(rk)
    //log.Println("[pbkdf2_validate] result: ", result)
    if result == parts[2] {
        return true, nil
    } else {
        return false, nil
    }
    return false, nil
}

func checkPasswordHash(salthash string, password string) bool {
    isok, err := pbkdf2_validate(password, salthash)
    if !isok || err != nil {
        log.Printf("[checkPasswordHash] err=%s, invalid password=%s", err, password)
        return false
    }

    return true
}

func genPasswordHash(password string) (salthash string, err error) {
    method := "sha1"
    iterations := kDefaultIter
    salt := genSalt(8)
    rk := pbkdf2_hash(method, []byte(password), []byte(salt), iterations)
    result := hex.EncodeToString(rk)

    if len(result) <= 0 {
        err = ErrHashPasswordFailure
        return
    }

    salthash = fmt.Sprintf("pbkdf2:%s:%d$%s$%s", method, iterations, salt, result)
    //log.Printf("[genPasswordHash] password=%s, salthash=%s", password, salthash)
    return
}

