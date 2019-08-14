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
	ErrConferenceInvalidArgument = errors.New("conference invalid argument")

	ErrConferenceWrongPassword = errors.New("conference had wrong password")
	ErrConferenceClosed        = errors.New("conference had been closed")
	ErrConferenceEnded         = errors.New("conference had been ended")
	ErrConferenceReachMaxSize  = errors.New("conference had reached max size")
	ErrConferenceNotCreator    = errors.New("conference not creator")
)
