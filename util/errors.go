package util;

import (
    "errors"
)

var (
    ErrFailed = errors.New("failed")

    ErrInvalidRequestArgs = errors.New("Invalid request args")

    ErrHashPasswordFailure = errors.New("Fail to hash password")
    ErrClientNotFound = errors.New("Client Not Found")

    ErrServerFault = errors.New("server fault")
    ErrInvalidArgument = errors.New("invalid argument")

    ErrUserExist = errors.New("user has been exist")
    ErrUserNotExist = errors.New("user not exist")
    ErrUserCreate = errors.New("user create failed")
    ErrUpdatePassword = errors.New("update password failed")

    ErrInvalidPassword = errors.New("invalid password")

    // ErrInvalidCredential is returned when the auth token does not authenticate correctly.
    ErrInvalidCredential = errors.New("invalid authorization credential")

    // ErrAuthenticationFailure returned when authentication failure to be presented to agent.
    ErrAuthenticationFailure = errors.New("authentication failure")
)

