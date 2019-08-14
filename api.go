package main

import (
	"log"
	"net/http"
)

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[api], begin", r)
	switch r.Method {
	case http.MethodPost:
	case http.MethodGet:
	default:
	}
}
