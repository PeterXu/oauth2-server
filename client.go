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
	"gopkg.in/oauth2.v3/utils/uuid"

	"github.com/PeterXu/oauth2-server/mongo"
	"github.com/PeterXu/oauth2-server/redis"
	"github.com/PeterXu/oauth2-server/util"
)

var (
	jsonMarshal   = jsoniter.Marshal
	jsonUnmarshal = jsoniter.Unmarshal
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
	Cid      int           `json:"cid"`
	Title    string        `json:"title"`
	Creator  string        `json:"creator"`
	Password string        `json:"password"`
	MaxSize  int           `json:"maxSize"`
	Start    time.Time     `json:"start"`
	Period   bool          `json:"period"`
	Duration time.Duration `json:"duration"`

	// dynamic status
	HostId  string          `json:"hostid"`
	Closed  bool            `json:"closed"`
	Rosters map[string]bool `json:"rosters"`
	Servers []string        `json:"servers"`
}

func NewConference(title, creator, password string) *Conference {
	return &Conference{
		Cid:      0,
		Title:    title,
		Creator:  creator,
		Password: password,
		MaxSize:  0,
		Start:    time.Now(),
		Period:   false,
		Duration: 0,
		HostId:   creator,
		Closed:   false,
		Rosters:  make(map[string]bool),
	}
}

func genInt3(n0, delta0 int, n2, delta2 int, special bool) int {
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

	// check special
	for {
		if !special {
			break
		}
		// one is 0
		if (nn[0] & nn[1] & nn[2]) == 0 {
			break
		}
		// equidifferent
		if (nn[0] - nn[1]) == (nn[1] - nn[2]) {
			break
		}
		// equal proportion
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
	rand.Seed(time.Now().UnixNano())
	p0 := genInt3(8, 1, 9, 0, true)
	rd := rand.Intn(2)
	p1 := genInt3(9, 0, 9, 0, (rd == 0))
	p2 := genInt3(9, 0, 8, 1, (rd == 1))
	return (((p0*1000 + p1) * 1000) + p2)
}

func (s *TokenStoreX) wrapperKey(id int) string {
	return fmt.Sprintf("%s-%d", "oauth2-conference", id)
}

func (s *TokenStoreX) wrapperKey2(id string) string {
	return fmt.Sprintf("%s-%s", "oauth2-uid-tmp", id)
}

func (s *TokenStoreX) createConference(info *Conference) error {
	ct := time.Now()
	for {
		cid := genConferenceId()
		if conf, err := s._checkConference(cid); err != nil {
			info.Cid = cid
			break
		} else {
			duration := ct.Sub(conf.Start)
			if duration > kTimeoutDuration {
				info.Cid = cid
				break
			}
		}
	}

	info.HostId = info.Creator
	if info.MaxSize < kDefaultUserSize {
		info.MaxSize = kDefaultUserSize
	}
	if info.Duration < kDefaultDuration {
		info.Duration = kDefaultDuration
	}
	if len(info.Title) == 0 || len(info.Title) >= 512 || len(info.Creator) == 0 {
		return util.ErrConferenceInvalidArgument
	}
	if len(info.Servers) == 0 {
		info.Servers = append(info.Servers, "rtc.zenvv.com")
	}

	return s._updateConference(info)
}

func (s *TokenStoreX) joinConference(cid int, fromId, password string) (*Conference, error) {
	if conf, err := s._checkConference(cid); err == nil {
		if conf.Closed {
			return nil, util.ErrConferenceClosed
		}

		if conf.MaxSize < kDefaultUserSize {
			conf.MaxSize = kDefaultUserSize
		}
		if conf.Duration < kDefaultDuration {
			conf.Duration = kDefaultDuration
		}
		ct := time.Now()
		duration := ct.Sub(conf.Start)
		if duration >= conf.Duration {
			return nil, util.ErrConferenceEnded
		}

		if password != conf.Password {
			return nil, util.ErrConferenceWrongPassword
		}

		liveSize := 0
		for _, had := range conf.Rosters {
			if had {
				liveSize += 1
			}
		}
		if liveSize >= conf.MaxSize {
			return nil, util.ErrConferenceReachMaxSize
		}

		conf.Rosters[fromId] = true
		if err := s._updateConference(conf); err != nil {
			return nil, err
		}
		return conf, nil
	}

	return nil, util.ErrConferenceNotExist
}

func (s *TokenStoreX) leaveConference(cid int, fromId string) error {
	if conf, err := s._checkConference(cid); err == nil {
		if _, ok := conf.Rosters[fromId]; ok {
			conf.Rosters[fromId] = false
		}
		if fromId == conf.Creator {
			conf.Closed = true
		} else {
			if fromId == conf.HostId {
				conf.HostId = conf.Creator
			}
		}
		return s._updateConference(conf)
	} else {
		return err
	}
}

func (s *TokenStoreX) updateConferenceHost(cid int, fromId string, hostId string) error {
	if conf, err := s._checkConference(cid); err == nil {
		if conf.Creator != fromId {
			return util.ErrConferenceNotCreator
		}
		conf.HostId = hostId
		return s._updateConference(conf)
	} else {
		return err
	}
}

func (s *TokenStoreX) _checkConference(cid int) (*Conference, error) {
	if s.rts != nil {
		result := s.rts.CLI().Get(s.wrapperKey(cid))
		if buf, err := s.rts.ParseData(result); err == nil {
			if conf, err := s._parseConference(buf); err == nil {
				return conf, nil
			}
		}
	}
	return nil, util.ErrConferenceNotExist
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
		pipe.Set(s.wrapperKey(info.Cid), jv, 0)
		if _, err := pipe.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (s *TokenStoreX) _updateData(key string, value interface{}, duration time.Duration) error {
	if s.rts != nil {
		pipe := s.rts.CLI().TxPipeline()
		pipe.Set(key, value, duration)
		if _, err := pipe.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (s *TokenStoreX) _checkData(key string) error {
	if s.rts != nil {
		result := s.rts.CLI().Get(key)
		if buf, err := s.rts.ParseData(result); err == nil {
			if len(buf) > 0 {
				return nil // exist
			}
			return util.ErrNotExist
		}
	}
	return util.ErrNotDone
}

func (s *TokenStoreX) randUid() string {
	var uid string
	quit := false
	tickChan := time.NewTicker(time.Millisecond * 100).C
	for !quit {
		select {
		case <-tickChan:
			quit = true
			break
		default:
			uid = uuid.Must(uuid.NewRandom()).String()
			key := s.wrapperKey2(uid)
			if err := s._checkData(key); err != nil {
				if err := s._updateData(key, true, 0); err == nil {
					quit = true
					break
				}
			}
		}
	}
	return uid
}

func (s *TokenStoreX) checkUid(uid string) error {
	key := s.wrapperKey2(uid)
	return s._checkData(key)
}
