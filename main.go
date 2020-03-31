package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port int `envconfig:"port" default:"8080"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var exch = NewExchanger()

func main() {
	var c Config
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/ws", ws)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", c.Port), nil))
}

func ws(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("upgrade error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	id, err := uuid.NewUUID()
	if err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, []byte("could not create UUID"))
		fmt.Println("could not create UUID")
		return
	}
	defer fmt.Printf("connection closed (%s)\n", id.String())
	fmt.Printf("connection opened (%s)\n", id.String())

	if err := exch.Exchange(r.Context(), id, conn); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return
}
