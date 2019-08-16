package main

import (
	"time"
)

const (
	kMinPasswordLength int = 6
	kMinUsernameLength int = 6

	kDefaultClientID     string = "defaultID"
	kDefaultClientSecret string = "defaultSecret"
	kDefaultClientDomain string = "http://localhost"

	kDefaultConfig string = "server.toml"

	// conference
	kDefaultUserSize = 3
	kTimeoutDuration = 120 * 12 * 30 * 24 * time.Hour // two JIA ZI
	kDefaultDuration = 45 * time.Minute
)
