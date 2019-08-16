package util

import (
	"errors"
)

// oauth2 token/code
var (
	ErrFailed = errors.New("failed")

	ErrInvalidRequestArgs = errors.New("Invalid request args")

	ErrHashPasswordFailure = errors.New("Fail to hash password")
	ErrClientNotFound      = errors.New("Client Not Found")

	ErrServerFault     = errors.New("server fault")
	ErrInvalidArgument = errors.New("invalid argument")

	ErrUserExist      = errors.New("user has been exist")
	ErrUserNotExist   = errors.New("user not exist")
	ErrUserCreate     = errors.New("user create failed")
	ErrUpdatePassword = errors.New("update password failed")

	ErrInvalidPassword = errors.New("invalid password")

	// ErrInvalidCredential is returned when the auth token does not authenticate correctly.
	ErrInvalidCredential = errors.New("invalid authorization credential")

	// ErrAuthenticationFailure returned when authentication failure to be presented to agent.
	ErrAuthenticationFailure = errors.New("authentication failure")
)

// oauth2 conference
var (
	ErrNotDone  = errors.New("not done")
	ErrNotExist = errors.New("not exist")

	ErrConferenceNoAccessToken      = errors.New("conference no access token")
	ErrConferenceInvalidAccessToken = errors.New("conference invalid access token")
	ErrConferenceNoPriviledge       = errors.New("conference no priviledge")

	ErrConferenceInvalidRequest  = errors.New("conference invalid request")
	ErrConferenceInvalidArgument = errors.New("conference invalid argument")

	ErrConferenceNotExist      = errors.New("conference not exist")
	ErrConferenceWrongPassword = errors.New("conference wrong password")
	ErrConferenceClosed        = errors.New("conference closed")
	ErrConferenceEnded         = errors.New("conference ended")
	ErrConferenceReachMaxSize  = errors.New("conference reached max size")
	ErrConferenceNotCreator    = errors.New("conference not creator")
)
