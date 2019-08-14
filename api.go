package main

import (
	"log"
	"net/http"
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
	if strings.HasPrefix(r.URL.Path, kConferenceApiV1) {
		cid := 0
		fromId := ""
		if strings.Compare(r.URL.Path, kConferenceApiV1Create) == 0 {
			conf := NewConference("Testing Conference", "test521", "")
			if err := gg.Store.createConference(conf); err != nil {
				ResponseErrorWithJson(w, err)
			} else {
				// TODO: response conference id
				ResponseSuccessWithJson(w)
			}
		} else if strings.Compare(r.URL.Path, kConferenceApiV1Join) == 0 {
			if err := gg.Store.joinConference(cid, fromId); err != nil {
				ResponseErrorWithJson(w, err)
			} else {
				// TODO: response media server
				ResponseSuccessWithJson(w)
			}
		} else if strings.Compare(r.URL.Path, kConferenceApiV1Update) == 0 {
			hostId := ""
			if err := gg.Store.updateConferenceHost(cid, fromId, hostId); err != nil {
				ResponseErrorWithJson(w, err)
			} else {
				// TODO: response media server
				ResponseSuccessWithJson(w)
			}
		} else if strings.Compare(r.URL.Path, kConferenceApiV1Leave) == 0 {
			if err := gg.Store.leaveConference(cid, fromId); err != nil {
				ResponseErrorWithJson(w, err)
			} else {
				// TODO: response media server
				ResponseSuccessWithJson(w)
			}
		} else {
			ResponseErrorWithJson(w, util.ErrConferenceInvalidRequest)
		}
	} else {
		ResponseErrorWithJson(w, util.ErrConferenceInvalidRequest)
	}
}
