package engine

import "errors"

var ErrSessionAlreadyInitialized = errors.New("session is already initialized")