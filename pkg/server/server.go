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

var (
	cl logpkg.CustomLogger

	entering, leaving chan *client
	messages          chan string
	done              chan struct{}
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
//
//	client entered
//	client left
//	message broadcasting
func broadcaster() {
	const (
		serverlTimeoutInMinutes = 10

		sendSleepTimeInMillis = 50
	)

	var (
		clients = make(map[client]bool)
		ticker  = time.NewTicker(3500 * time.Minute)
	)

	cl.Println(cfg.LogSplitter)
	cl.Println(cfg.SessionStartedMsg)

	defer func() {
		ticker.Stop()

		if err := cl.Close(); err != nil {
			panic(err)
		}
	}()

outer:
	for {
		select {

		case newClient := <-entering:
			newClient.c <- announceAllClients(clients)
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

	cl.Println(cfg.SessionCloseddMsg)
	cl.Println(cfg.LogSplitter)
	cl.Println()
}

// handleConn serves all the client side work
func handleConn(conn net.Conn) {
	const (
		msgTimeoutInMin = 2
	)

	var (
		client = NewClient(conn)

		ticker      = time.NewTicker(msgTimeoutInMin * time.Minute)
		messageSent = make(chan struct{})
	)

	// Sets a client's name
	client.setName()

	// Defers the ticker stop and closing actions of the client
	defer func() {
		ticker.Stop()
		close(messageSent)
	}()

welcomeuser:
	for {
		select {
		case messages <- fmt.Sprintf(cfg.WelcomeMsgFmt, client.name):

		case <-done:
			break welcomeuser
		}
	}

	// Starts reading messages from the clients' pool
	go func() {
		client.getChatMessages()
	}()

sendclientinfo:
	for {
		select {
		case client.c <- fmt.Sprintf("YOU ARE: %s", client.name):

		case entering <- client:

		case <-done:
			break sendclientinfo
		}
	}

	// Starts reading messages from a client
	go func() {
		scanner := bufio.NewScanner(conn)

	clientmsgcheck:
		for scanner.Scan() {
			select {

			case messages <- fmt.Sprintf("%s: %s", client.name, scanner.Text()):

			case messageSent <- struct{}{}:

			case <-done:
				break clientmsgcheck
			}
		}
	}()

poolmsgscheck:
	for {
		select {
		case <-ticker.C:
			break poolmsgscheck

		case <-messageSent:
			ticker.Reset(msgTimeoutInMin * time.Minute)

		case <-done:
			break poolmsgscheck
		}
	}
	ticker.Stop()

	messages <- fmt.Sprintf(cfg.GoodbyeMsgFmt, client.name)

	leaving <- client
}

// announceAllClients puts all the usernames into a buffer created and returns its string representation
func announceAllClients(clients map[client]bool) string {
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
