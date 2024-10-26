package network

import "errors"

var (
	ErrTooManyConnection   = errors.New("too many connection")
	ErrConnectionOpened    = errors.New("connection is opened")
	ErrConnectionHanged    = errors.New("connection is hanged")
	ErrConnectionClosed    = errors.New("connection is closed")
	ErrConnectionNotOpened = errors.New("connection is not opened")
	ErrConnectionNotHanged = errors.New("connection is not hanged")
)
