package main

import (
	//"log"
	"github.com/BurntSushi/toml"
	"github.com/PeterXu/oauth2-server/util"
)

func NewConfig(fname string) (conf Config, err error) {
	_, err = toml.DecodeFile(fname, &conf)
	return
}

type Config struct {
	Title   string
	Flow    string
	Listen  listenInfo
	Clients map[string]clientInfo
	Store   storeInfo
	Db      dbInfo
	Service serviceInfo
}

func (c *Config) GetClientByID(id string) (cinfo clientInfo, err error) {
	//log.Printf("clients: ", c.Clients)
	for _, cli := range c.Clients {
		if cli.Id == id {
			cinfo = cli
			return
		}
	}

	err = util.ErrClientNotFound
	return
}

type listenInfo struct {
	Host string
	Port int
}

type clientInfo struct {
	Id     string
	Secret string
	Domain string
	Grants []string //> grant types
	Scopes []string //> self-defined scopes
}

type storeInfo struct {
	// redis or mongo
	Engine string
	Host   string
	Port   int
	Db     string //> only for mongo
}

type dbInfo struct {
	// mysql
	Engine string
	// (a) oauth:oauth@/oauth;
	// (b) oauth:oauth@tcp(127.0.0.1:3306)/oauth
	Connection string
}

type serviceInfo struct {
	Enable bool
	Host   string
	Port   int
}
