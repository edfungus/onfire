package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port int `envconfig:"port" default:"8080"`
}

var upgrader = websocket.Upgrader{}

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
		fmt.Errorf("upgrade error: %+v", err)
	}
	defer conn.Close()
	defer fmt.Println("connection closed")
	fmt.Println("connection opened")
	for {
		mt, mb, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("error reading message", err)
			return
		}
		fmt.Println("got message: %s", mb)
		err = conn.WriteMessage(mt, mb)
		if err != nil {
			fmt.Println("error sending message", err)
			return
		}
	}
}
