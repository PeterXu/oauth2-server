package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PeterXu/oauth2-server/util"
)

const (
	kConferenceApiV1 = "/api/v1/"

	// Conference api
	kConferenceApiV1Create = kConferenceApiV1 + "create"
	kConferenceApiV1Join   = kConferenceApiV1 + "join"
	kConferenceApiV1Update = kConferenceApiV1 + "update"
	kConferenceApiV1Leave  = kConferenceApiV1 + "leave"
)

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[api], begin", r.URL)
	token := strings.TrimSpace(r.FormValue("token"))
	if len(token) > 0 {
		return
	}

	if strings.HasPrefix(r.URL.Path, kConferenceApiV1) {
		fromId := strings.TrimSpace(r.FormValue("fromid"))
		if len(fromId) == 0 {
			Response400ErrorWithJson(w, util.ErrConferenceInvalidArgument)
			return
		}

		if strings.Compare(r.URL.Path, kConferenceApiV1Create) == 0 {
			title := strings.TrimSpace(r.FormValue("title"))
			password := strings.TrimSpace(r.FormValue("password"))

			conf := NewConference(title, fromId, password)
			if err := gg.Store.createConference(conf); err != nil {
				Response400ErrorWithJson(w, err)
			} else {
				Response200DataWithJson(w, map[string]interface{}{
					"cid":     conf.Cid,
					"servers": conf.Servers,
					"token":   "conf token",
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
					"token":   "conf token",
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
