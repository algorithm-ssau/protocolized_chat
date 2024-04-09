package main

import (
	"fmt"
	. "protocolized_chat/pkg/server"
	"runtime"
)

func main() {
	defer fmt.Println(runtime.NumGoroutine())
	ChatServer()

}
