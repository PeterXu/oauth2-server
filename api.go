package main

import (
	"log"
	"net/http"
)

// Conference api:
//	1). api/v1/join
//  2). api/v1/join#cid
//  3). api/v1/update?hostid=
//	3). api/v1/leave#cid
func ApiHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[api], begin", r)
	switch r.Method {
	case http.MethodPost:
	case http.MethodGet:
	default:
	}
}
