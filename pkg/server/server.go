package server

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"time"

	cfg "protocolized_chat/pkg/config"
	logpkg "protocolized_chat/pkg/log"
)

// init makes package's preparation
func init() {
	cl = *logpkg.GetLogger()

	entering = make(chan *client)
	leaving = make(chan *client)

	messages = make(chan string)

	done = make(chan struct{})
}

// ChatServer creates the default server on the 8080 port and listens incoming connections
func ChatServer() {

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		broadcaster()
	}()

	go func() {
		<-done
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			break
		}

		go handleConn(conn)

	}

}

// broadcaster is a monitor functions that handles all the actions in the clients' poll such as:
// - client entered
// - client left
// - message broadcasting
func broadcaster() {
	const (
		serverlTimeoutInMinutes = 10
		sendSleepTimeInMillis   = 50
	)

	var (
		clients = make(map[client]bool)
		ticker  = time.NewTicker(serverlTimeoutInMinutes * time.Minute)
	)

	cl.Println(cfg.LogSplitter)
	cl.Println(cfg.SessionStartedMsg)

	defer func() {
		ticker.Stop()

		if err := cl.Close(); err != nil {
			panic(err)
		}

		cl.Println(cfg.SessionCloseddMsg)
		cl.Println(cfg.LogSplitter)
		cl.Println()
	}()

outer:
	for {
		select {

		case newClient := <-entering:
			newClient.c <- getClientsList(clients)
			clients[*newClient] = true

		case leftClient := <-leaving:
			delete(clients, *leftClient)
			leftClient.Close()

		case msg := <-messages:
			cl.Println(msg)

			for client := range clients {
			inner:
				for {
					select {
					case client.c <- msg:
						break inner
					default:
						time.Sleep(sendSleepTimeInMillis * time.Millisecond)
					}
				}
			}

		case <-ticker.C:
			close(done)
			break outer
		}

	}

}

// handleConn serves all the client side work
func handleConn(conn net.Conn) {
	const (
		msgTimeoutInMin = 2
	)

	var (
		ct          = NewCustomTicker(msgTimeoutInMin)
		messageSent = make(chan struct{})

		client = NewClient(conn)

		signal <-chan struct{}
	)

	// Sets a client's name
	client.setName()

	signal = welcome(client)
	<-signal

	// Defers the ticker stop and closing actions of the client
	defer func() {
		ct.ticker.Stop()
		close(messageSent)
	}()

	// Starts reading messages from the clients' pool
	go func() {
		client.dumpChatMessages()
	}()

	signal = sendClientInfo(client)
	<-signal

	// Starts reading messages from a client
	go checkClientMessages(client, conn, messageSent)

	checkPoolMessages(ct, messageSent)
	ct.ticker.Stop()

	messages <- fmt.Sprintf(cfg.GoodbyeMsgFmt, client.name)

	leaving <- client
}

func welcome(client *client) <-chan struct{} {

	var (
		innerDone      = make(chan struct{})
		messagesCopied = messages
	)

	go func() {
	welcomeuser:
		for i := 0; i < 1; i++ {
			select {
			case messagesCopied <- fmt.Sprintf(cfg.WelcomeMsgFmt, client.name):
				messagesCopied = nil
				break
			case <-innerDone:
				break welcomeuser
			}
		}

		close(innerDone)
	}()

	return innerDone
}

func sendClientInfo(client *client) <-chan struct{} {

	var (
		innerDone        = make(chan struct{})
		clientChanCopied = client.c
		enteringCopied   = entering
	)

	go func() {
		for i := 0; i < 2; i++ {
			select {
			case clientChanCopied <- fmt.Sprintf("YOU ARE: %s", client.name):
				clientChanCopied = nil

			case enteringCopied <- client:
				enteringCopied = nil
				close(innerDone)

			case <-innerDone:
				return
			}
		}
	}()

	return innerDone
}

func checkClientMessages(client *client, conn net.Conn, messageSent chan struct{}) {
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		select {

		case messages <- fmt.Sprintf("%s: %s", client.name, scanner.Text()):
		case messageSent <- struct{}{}:

		case <-done:
			return
		}
	}
}

func checkPoolMessages(ct *CustomTicker, messageSent chan struct{}) {
	for {
		select {
		case <-ct.ticker.C:
			return

		case <-messageSent:
			ct.ticker.Reset(ct.msgTimeoutInMin * time.Minute)

		case <-done:
			return
		}
	}
}

// getClientsList puts all the usernames into a buffer created and returns its string representation
func getClientsList(clients map[client]bool) string {
	var (
		currentClients bytes.Buffer
	)

	currentClients.WriteString("USERS ONLINE:")
	currentClients.WriteByte('\n')

	for client := range clients {
		select {
		case <-done:
			return cfg.ServerClosedMsg

		default:
			currentClients.WriteString(client.name)
			currentClients.WriteByte('\n')
		}

	}

	return currentClients.String()
}
