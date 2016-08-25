package util;

import (
    "errors"
)

var (
    ErrInvalidPassword = errors.New("invalid password")

    // ErrInvalidCredential is returned when the auth token does not authenticate correctly.
    ErrInvalidCredential = errors.New("invalid authorization credential")

    // ErrAuthenticationFailure returned when authentication failure to be presented to agent.
    ErrAuthenticationFailure = errors.New("authentication failure")
)

