package common

import "github.com/pkg/errors"

var (
    ERR_LOCKED = errors.New("LOCK IS REQUIRED")
)
