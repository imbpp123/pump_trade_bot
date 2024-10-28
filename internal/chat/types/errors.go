package types

import "errors"

var (
	ErrChatNoMessage = errors.New("no message in update")
	ErrChatStarted   = errors.New("chat already started")
)
