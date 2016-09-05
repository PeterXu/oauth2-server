package main

import (
    "log"
    "github.com/BurntSushi/toml"
)

func NewConfig(fname string) (conf Config, err error) {
    if _, err = toml.DecodeFile(fname, &conf); err != nil {
        log.Fatal(err)
        return
    }
    log.Printf("Title: %s", conf.Title)
    return
}


type Config struct {
    Title string
    Listen listenInfo
    Clients map[string]clientInfo
    Store storeInfo
    Db dbInfo
}

type listenInfo struct {
    Host string
    Port int
}

type clientInfo struct {
    Id string
    Secret string
    Domain string
}

type storeInfo struct {
    Engine string
    Host string
    Port int
    Db string
}

type dbInfo struct {
    Engine string
    Connection string
}

