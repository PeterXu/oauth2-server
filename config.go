package main

import (
    "log"
    "github.com/BurntSushi/toml"
    "./util"
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

func (c *Config)GetClientByID(id string) (cinfo clientInfo, err error) {
    err = util.ErrClientNotFound
    log.Printf("clients: ", c.Clients)
    for _, cli := range(c.Clients) {
        if cli.Id == id {
            cinfo = cli
            err = nil
            return
        }
    }
    return
}

type listenInfo struct {
    Host string
    Port int
}

type clientInfo struct {
    Id string
    Secret string
    Domain string
    Grants []string  //> grant types
}

type storeInfo struct {
    // redis or mongo
    Engine string
    Host string
    Port int
    Db string       //> only for mongo
}

type dbInfo struct {
    // mysql
    Engine string
    // (a) oauth:oauth@/oauth; 
    // (b) oauth:oauth@tcp(127.0.0.1:3306)/oauth
    Connection string
}

