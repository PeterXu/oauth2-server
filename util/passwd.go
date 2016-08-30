package util

import (
    //"log"
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

func hashPasswordWithSalt(salt, password string, cnt int) string {
    sha := sha256.Sum256([]byte(salt+password)) // 32bytes
    for k := 1; k < cnt; k++ {
        sha = sha256.Sum256(sha[:])
    }
    return genBase64(sha[:])
}

func hashPasswordNoSalt(password string, cnt int) (salt, hash string) {
    salt = genSalt()
    hash = hashPasswordWithSalt(salt, password, cnt)
    //log.Printf("[hashPasswordNoSalt] salt=%s, hash=%s, pass=%s", salt, hash, password)
    return
}

