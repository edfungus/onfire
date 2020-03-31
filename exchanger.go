package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Exchanger takes two ws connections and exchanges WebRTC connection information
type Exchanger struct {
	queue chan *connection
	hash  map[uuid.UUID]*connection
}

type connection struct {
	id         uuid.UUID
	connection *websocket.Conn
	done       chan bool
}

func NewExchanger() *Exchanger {
	return &Exchanger{
		queue: make(chan *connection),
		hash:  map[uuid.UUID]*connection{},
	}
}

// Exchange shares the WebRTC details between two connections.
// This is a blocking call until the exchange is finished
func (e *Exchanger) Exchange(ctx context.Context, id uuid.UUID, conn *websocket.Conn) error {
	c := &connection{
		id:         id,
		connection: conn,
		done:       make(chan bool),
	}
	// Add connection to hash so we can set it to nil if the connection terminates
	e.hash[c.id] = c
	defer func() {
		// Clears the hash and prevent a closed connection to be read off the queue
		e.hash[c.id] = nil
	}()
	select {
	case e.queue <- c:
		// Wait for the connection to be process or exit because connection context is closed
		select {
		case <-c.done:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("context cancelled")
		}
	case xc := <-e.queue:
		if err := negotiate(c, xc); err != nil {
			return nil
		}
		// Unblock Exchange() call of the other connection
		close(xc.done)
	case <-ctx.Done():
		return fmt.Errorf("context cancelled")
	}
	return nil
}

func negotiate(c1 *connection, c2 *connection) error {
	c1.connection.WriteMessage(websocket.TextMessage, []byte(c2.id.String()))
	c2.connection.WriteMessage(websocket.TextMessage, []byte(c1.id.String()))
	return nil
}
