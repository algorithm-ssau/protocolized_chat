package server

import (
	"bufio"
	"fmt"
	"net"
)

const (
	nameRequestMsg = "ENTER YOUR NAME: "
)

type client struct {
	name string
	conn net.Conn
	c    chan string
}

// NewClient returns a new client based on conn
func NewClient(conn net.Conn) *client {
	return &client{conn: conn, c: make(chan string)}
}

// setName sets a name of a client from a user input
func (client *client) setName() {
	var (
		scanner = bufio.NewScanner(client.conn)
	)

	fmt.Fprint(client.conn, "\n")
	fmt.Fprint(client.conn, nameRequestMsg)

	for {
		select {
		case <-done:
			return

		default:
			if scanner.Scan() {
				select {
				case <-done:
					return

				default:
					client.name = scanner.Text()
					fmt.Println(client.name)
					return
				}
			} else if err := scanner.Err(); err != nil {
				cl.Println(fmt.Sprintf("while client with address %s: %s", client.conn.LocalAddr(), err))
			}

		}
	}

}

// dumpChatMessages fetches the message from a client's pool and prints it in the client side
func (client *client) dumpChatMessages() {
	for msg := range client.c {
		select {
		case <-done:
			return

		default:
			fmt.Fprintln(client.conn, msg)

		}

	}
}

// Closes the client's resourses:
// - conn
// - c
func (client *client) Close() {
	close(client.c)
	client.conn.Close()

}
