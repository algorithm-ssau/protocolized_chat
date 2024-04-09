package server

import (
	"bufio"
	"fmt"
	"net"
)

const (
	clientBufferSize = 16
	nameRequestMsg   = "ENTER YOUR NAME: "
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
					return
				}
			} else if err := scanner.Err(); err != nil {
				cl.Println(fmt.Sprintf("while client with address %s: %s", client.conn.LocalAddr(), err))
			}

		}
	}

}

// getChatMessages fetches the message from a client's pool and prints it in the client side
func (client *client) getChatMessages() {
	for msg := range client.c {
		select {
		case <-done:
			return

		default:
			fmt.Fprintln(client.conn, msg)

		}

	}
}

// Closes the client's resourses such as conn and c
func (client *client) Close() {
	client.conn.Close()
	close(client.c)
}
