package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/models"

	"github.com/PeterXu/oauth2-server/mongo"
	"github.com/PeterXu/oauth2-server/redis"
	"github.com/PeterXu/oauth2-server/util"
)

var (
	jsonMarshal   = jsoniter.Marshal
	jsonUnmarshal = jsoniter.Unmarshal
)

const (
	kDefaultClientID     string = "defaultID"
	kDefaultClientSecret string = "defaultSecret"
	kDefaultClientDomain string = "http://localhost"
)

type MyClientStore struct {
	data map[string]*models.Client
}

func (ts *MyClientStore) GetByID(id string) (cli oauth2.ClientInfo, err error) {
	err = util.ErrClientNotFound
	if c, ok := ts.data[id]; ok {
		cli = c
		err = nil
	}
	return
}

func NewMyClientStore(clients map[string]clientInfo) oauth2.ClientStore {
	data := map[string]*models.Client{
		kDefaultClientID: &models.Client{
			ID:     kDefaultClientID,
			Secret: kDefaultClientSecret,
			Domain: kDefaultClientDomain,
		},
	}
	for _, cli := range clients {
		data[cli.Id] = &models.Client{
			ID:     cli.Id,
			Secret: cli.Secret,
			Domain: cli.Domain,
		}
	}
	//log.Printf("[NewMyClientStore] clients: ", data, len(data))
	return &MyClientStore{
		data: data,
	}
}

func NewTokenStore(sinfo storeInfo) (store *TokenStoreX, err error) {
	address := sinfo.Host + ":" + strconv.Itoa(sinfo.Port)
	switch sinfo.Engine {
	case "mongo":
		mstore := mongo.NewTokenStore(mongo.NewConfig(
			"mongodb://"+address,
			sinfo.Db,
		), mongo.NewDefaultTokenConfig())
		store = &TokenStoreX{mstore, mstore, nil}
	case "redis":
		options := &redis.Options{
			Addr: address,
		}
		rstore := redis.NewRedisStore(options)
		store = &TokenStoreX{rstore, nil, rstore}
	default:
		log.Println("[NewTokenStore] Unsupported storage engine: ", sinfo.Engine)
		return
	}

	return
}

type TokenStoreX struct {
	oauth2.TokenStore
	mts *mongo.TokenStore
	rts *redis.TokenStore
}

type Conference struct {
	cid      int
	title    string
	creator  string
	password string
	maxSize  int
	start    time.Time
	duration time.Duration

	// dynamic status
	hostId  string
	closed  bool
	rosters map[string]bool
}

func NewConference(title, creator, password string) *Conference {
	return &Conference{
		cid:      0,
		title:    title,
		creator:  creator,
		password: password,
		hostId:   creator,
		closed:   false,
		rosters:  make(map[string]bool),
	}
}

func genInt3(n0, delta0 int, n2, delta2 int) int {
	nn := make([]int, 3)
	nn[0] = rand.Intn(n0) + delta0
	if nn[0] == 0 {
		nn[1] = rand.Intn(8) + 1
		nn[2] = rand.Intn(8) + 1
	} else {
		nn[1] = rand.Intn(9)
		if nn[1] == 0 {
			nn[2] = rand.Intn(8) + 1
		} else {
			nn[2] = rand.Intn(n2) + delta2
		}
	}
	for {
		// one is 0
		if (nn[0] & nn[1] & nn[2]) == 0 {
			break
		}
		if (nn[0] - nn[1]) == (nn[1] - nn[2]) {
			break
		}
		if (nn[0] / nn[1]) == (nn[1] / nn[2]) {
			break
		}
		idx := rand.Intn(2)
		nn[2] = nn[idx]
		break
	}
	return (nn[0]*100 + nn[1]*10 + nn[2])
}

func genConferenceId() int {
	p0 := genInt3(8, 1, 9, 0)
	p1 := genInt3(9, 0, 9, 0)
	p2 := genInt3(9, 0, 8, 1)
	return (((p0*1000 + p1) * 1000) + p2)
}

func (s *TokenStoreX) wrapperKey(id int) string {
	return fmt.Sprintf("%s-%d", "oauth2-conference", id)
}

func (s *TokenStoreX) createConference(info *Conference) error {
	ct := time.Now()
	for {
		cid := genConferenceId()
		if conf, err := s._checkConference(cid); err != nil {
			info.cid = cid
			break
		} else {
			duration := ct.Sub(conf.start)
			const kTimeoutDuration = 60 * 12 * 30 * 24 * time.Hour // 12 years
			if duration > kTimeoutDuration {
				info.cid = cid
				break
			}
		}
	}

	info.start = ct
	if info.maxSize >= 0 && info.maxSize < 3 {
		info.maxSize = 3
	}
	if len(info.title) == 0 || len(info.title) >= 512 || len(info.creator) == 0 {
		return util.ErrConferenceInvalidArgument
	}

	return s._updateConference(info)
}

func (s *TokenStoreX) joinConference(cid int, fromId string) error {
	if conf, err := s._checkConference(cid); err == nil {
		if conf.closed {
			return util.ErrConferenceClosed
		}

		ct := time.Now()
		duration := ct.Sub(conf.start)
		if duration >= conf.duration {
			return util.ErrConferenceEnded
		}

		if conf.password != conf.password {
			return util.ErrConferenceWrongPassword
		}

		liveSize := 0
		for _, had := range conf.rosters {
			if had {
				liveSize += 1
			}
		}
		if conf.maxSize >= 0 && liveSize >= conf.maxSize {
			return util.ErrConferenceReachMaxSize
		}

		conf.rosters[fromId] = true
		return nil
	}

	return util.ErrConferenceNotExist
}

func (s *TokenStoreX) leaveConference(cid int, fromId string) error {
	if conf, err := s._checkConference(cid); err == nil {
		if _, ok := conf.rosters[fromId]; ok {
			conf.rosters[fromId] = false
		}
		if fromId == conf.creator {
			conf.closed = true
		} else {
			if fromId == conf.hostId {
				conf.hostId = conf.creator
			}
		}
		return s._updateConference(conf)
	} else {
		return err
	}
}

func (s *TokenStoreX) updateConferenceHost(cid int, fromId string, hostId string) error {
	if conf, err := s._checkConference(cid); err == nil {
		if conf.creator != fromId {
			return util.ErrConferenceNotCreator
		}
		conf.hostId = hostId
		return s._updateConference(conf)
	} else {
		return err
	}
}

func (s *TokenStoreX) _checkConference(cid int) (*Conference, error) {
	var err error
	if s.rts != nil {
		result := s.rts.CLI().Get(s.wrapperKey(cid))
		if buf, err := s.rts.ParseData(result); err == nil {
			if conf, err := s._parseConference(buf); err == nil {
				return conf, nil
			}
		}
	}
	return nil, err
}

func (s *TokenStoreX) _parseConference(buf []byte) (*Conference, error) {
	var info Conference
	if err := jsonUnmarshal(buf, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *TokenStoreX) _updateConference(info *Conference) error {
	jv, err := jsonMarshal(info)
	if err != nil {
		return err
	}

	if s.rts != nil {
		pipe := s.rts.CLI().TxPipeline()
		pipe.Set(s.wrapperKey(info.cid), jv, 0)
		if _, err := pipe.Exec(); err != nil {
			return err
		}
	}
	return nil
}
