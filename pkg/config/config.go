package config

import (
	"strings"
)

const (
	LogFilePath = "/Users/dmitriymamykin/Desktop/protocolized_chat/pkg/log/logs/log.txt"

	SessionStartedMsg = "SESSION STARTED"
	SessionCloseddMsg = "SESSION CLOSED"

	ServerClosedMsg = "SERVER CLOSE"

	WelcomeMsgFmt = "USER %s ENTERED THE CHAT"
	GoodbyeMsgFmt = "USER: %s LEFT THE CHAT"
)

var (
	LogSplitter = strings.Repeat("-", 128)
)
