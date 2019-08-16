package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PeterXu/oauth2-server/util"
)

const (
	kConferenceApiV1 = "/api/v1/"

	// Token/Auth api: ?token=xxx
	kConferenceApiV1Token = kConferenceApiV1 + "token"
	kConferenceApiV1Auth  = kConferenceApiV1 + "auth"

	// Conference api
	kConferenceApiV1Create = kConferenceApiV1 + "create"
	kConferenceApiV1Join   = kConferenceApiV1 + "join"
	kConferenceApiV1Update = kConferenceApiV1 + "update"
	kConferenceApiV1Leave  = kConferenceApiV1 + "leave"
)

// GET/POST: api/v1/action?token=xxx&cid=xxx
// Guest should request a temporal token.
func ApiHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[api], begin", r.URL)
	if strings.HasPrefix(r.URL.Path, kConferenceApiV1) {
		if strings.Compare(r.URL.Path, kConferenceApiV1Token) == 0 {
			u := new(url.URL)
			u.Parse(r.URL.String())
			u.Path = "/code"

			query := r.URL.Query()
			if query.Get("response_type") == "" {
				query.Set("response_type", "token")
			}
			if query.Get("exp") == "" {
				query.Set("exp", "3600")
			}
			u.RawQuery = query.Encode()

			//log.Println(u.String())
			w.Header().Set("Location", u.String())
			w.WriteHeader(http.StatusFound)
			return
		}

		if strings.Compare(r.URL.Path, kConferenceApiV1Auth) == 0 {
			u := new(url.URL)
			u.Parse(r.URL.String())
			u.Path = "/check"
			u.RawQuery = r.URL.RawQuery
			w.Header().Set("Location", u.String())
			w.WriteHeader(http.StatusFound)
			return
		}

		var fromId string
		token := strings.TrimSpace(r.FormValue("token"))
		if len(token) > 0 {
			if ti, err := gg.Server.Manager.LoadAccessToken(token); err == nil {
				fromId = ti.GetUserID()
			}
		}
		if len(fromId) == 0 {
			Response400ErrorWithJson(w, util.ErrConferenceInvalidAccessToken)
			return
		}

		if strings.Compare(r.URL.Path, kConferenceApiV1Create) == 0 {
			// Only normal user can create conference
			// Verify fromId which should not be an temporal user.
			if !gg.Users.CheckUserID(fromId, "") {
				Response400ErrorWithJson(w, util.ErrConferenceNoPriviledge)
				return
			}

			title := strings.TrimSpace(r.FormValue("title"))
			password := strings.TrimSpace(r.FormValue("password"))

			conf := NewConference(title, fromId, password)
			if err := gg.Store.createConference(conf); err != nil {
				Response400ErrorWithJson(w, err)
			} else {
				Response200DataWithJson(w, map[string]interface{}{
					"cid":     conf.Cid,
					"servers": conf.Servers,
				})
			}
			return
		}

		// cid required for other apis except kConferenceApiV1Create
		cid, err := strconv.Atoi(strings.TrimSpace(r.FormValue("cid")))
		if err != nil {
			Response400ErrorWithJson(w, util.ErrConferenceInvalidArgument)
			return
		}

		if strings.Compare(r.URL.Path, kConferenceApiV1Join) == 0 {
			password := strings.TrimSpace(r.FormValue("password"))
			if conf, err := gg.Store.joinConference(cid, fromId, password); err != nil {
				Response400ErrorWithJson(w, err)
			} else {
				Response200DataWithJson(w, map[string]interface{}{
					"cid":     conf.Cid,
					"servers": conf.Servers,
				})
			}
		} else if strings.Compare(r.URL.Path, kConferenceApiV1Update) == 0 {
			hostId := strings.TrimSpace(r.FormValue("hostid"))
			if err := gg.Store.updateConferenceHost(cid, fromId, hostId); err != nil {
				Response400ErrorWithJson(w, err)
			} else {
				Response200SuccessWithJson(w)
			}
		} else if strings.Compare(r.URL.Path, kConferenceApiV1Leave) == 0 {
			if err := gg.Store.leaveConference(cid, fromId); err != nil {
				Response400ErrorWithJson(w, err)
			} else {
				Response200SuccessWithJson(w)
			}
		} else {
			Response400ErrorWithJson(w, util.ErrConferenceInvalidRequest)
		}
	} else {
		Response400ErrorWithJson(w, util.ErrConferenceInvalidRequest)
	}
}
