package domain

import "errors"

var ErrSessionNotFound = errors.New("session not found")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrRegisterUser = errors.New("failed register user")
var ErrUserConflict = errors.New("a user with this email already exists")
var ErrTooManyRequests = errors.New("too many requests")
