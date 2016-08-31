package util

import (
    "log"
    "fmt"
    "strings"
    "strconv"
    //"bytes"
    "encoding/base64"
    crand "crypto/rand"
    //"crypto/md5"
    //"crypto/hmac"
    "crypto/sha256"
)

const kDefaultIter int = 10


func genRand(num int) []byte{
    b := make([]byte, num)
    _, err := crand.Read(b)
    if err != nil {
        return nil
    }
    return b
}

func genBase64(data []byte) string {
    return base64.StdEncoding.EncodeToString(data)
}

func genSalt() string {
    return genBase64(genRand(16))[0:8] // 8bytes
}

func hashPasswordWithSalt(crypt, salt, password string, cnt int) string {
    sha := sha256.Sum256([]byte(salt+password)) // 32bytes
    for k := 1; k < cnt; k++ {
        sha = sha256.Sum256(sha[:])
    }
    return genBase64(sha[:])
}

func hashPasswordNoSalt(password string, cnt int) (salt, hash string) {
    salt = genSalt()
    hash = hashPasswordWithSalt("sha256", salt, password, cnt)
    //log.Printf("[hashPasswordNoSalt] salt=%s, hash=%s, pass=%s", salt, hash, password)
    return
}

func genPasswordHash(password string) (salthash string, err error) {
    mark := "self"
    crypt := "sha256"
    cnt := kDefaultIter

    salt, hash := hashPasswordNoSalt(password, cnt)
    if len(salt) <= 0 || len(hash) <= 0 {
        err = ErrFailed
        return
    }

    salthash = fmt.Sprintf("%s:%s:%d$%s$%s", mark, crypt, cnt, salt, hash)
    return
}

func checkPasswordHash(salthash string, password string) bool {
    str := strings.Split(salthash, "$")
    if len(str) != 3 {
        log.Println("invalid (salt+hash): ", salthash)
        return false
    }

    mark := "self"
    crypt := "sha256"
    cnt := kDefaultIter
    salt := str[1]
    hash1 := str[2]

    // update default mark/crypt/cnt
    str2 := strings.Split(str[0], ":")
    if len(str2) == 3 { // self(custom) and werkzeug.security
        mark = str2[0]
        crypt = str2[1]
        cnt, _ = strconv.Atoi(str2[2])
    }

    hash2 := ""
    if mark == "self" {
        hash2 = hashPasswordWithSalt(crypt, salt, password, cnt)
    }

    if hash1 != hash2 {
        log.Println("invalid password: ", password)
        return false
    }
    return true
}
