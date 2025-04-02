package rabbitmq

import "errors"

var (
	ErrGetChannelTimeout = errors.New("get channel timeout")
	ErrConnectionClosed  = errors.New("connection closed")
	ErrChannelClosed     = errors.New("channel closed")
	ErrMaxRetriesReached = errors.New("max retries reached")
)
