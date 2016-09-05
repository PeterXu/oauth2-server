package main

import (
    "log"
    "strconv"
    "gopkg.in/oauth2.v3"
    "gopkg.in/oauth2.v3/models"
    "github.com/go-oauth2/mongo"
    "github.com/go-oauth2/redis"
)

type MyClientStore struct {
    data map[string]*models.Client
}

func (ts *MyClientStore) GetByID(id string) (cli oauth2.ClientInfo, err error) {
    if c, ok := ts.data[id]; ok {
        cli = c
    }
    return
}

func NewMyClientStore(clients map[string]clientInfo) oauth2.ClientStore {
    data := map[string]*models.Client{
        "1": &models.Client{
            ID:     "1",
            Secret: "11",
            Domain: "http://localhost",
        },
    }
    for _, cli := range clients {
        data[cli.Id] = &models.Client{
            ID:     cli.Id,
            Secret: cli.Secret,
            Domain: cli.Domain,
        }
    }
    log.Printf("[NewMyClientStore] clients: ", data, len(data))
    return &MyClientStore{
        data: data,
    }
}

func NewTokenStore(sinfo storeInfo) (store oauth2.TokenStore, err error){
    address := sinfo.Host + ":" + strconv.Itoa(sinfo.Port)
    switch sinfo.Engine {
    case "mongo":
        store, err = mongo.NewTokenStore(mongo.NewConfig(
            "mongodb://" + address,
            sinfo.Db,
        ))
    case "redis":
        store, err = redis.NewTokenStore(&redis.Config{
            Addr: address,
        })
    default:
        log.Println("[NewTokenStore] Unsupported storage engine: ", sinfo.Engine)
        return
    }

    return
}

