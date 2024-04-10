package server

import (
	logpkg "protocolized_chat/pkg/log"

	"time"
)

var (
	cl logpkg.CustomLogger

	entering, leaving chan *client
	messages          chan string
	done              chan struct{}
)

type CustomTicker struct {
	ticker          *time.Ticker
	msgTimeoutInMin time.Duration
}

func NewCustomTicker(timeoutInMin time.Duration) *CustomTicker {
	return &CustomTicker{ticker: time.NewTicker(timeoutInMin * time.Minute), msgTimeoutInMin: timeoutInMin}
}
