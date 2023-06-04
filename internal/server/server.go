package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"

	"skyline/config"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var connections map[uuid.UUID]*websocket.Conn

func ServerInit() {
	conf, _ := config.GetConfig()
	ConfigInit()
	log.Debugln("debug")
	connections = make(map[uuid.UUID]*websocket.Conn)
	http.HandleFunc("/socket", socketHandler)
	http.HandleFunc("/", home)
	err := http.ListenAndServe(conf.SERVER_HOST_PORT, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func ConfigInit() {
	username, err := user.Current()
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(fmt.Sprintf("/home/%s/skyline", username.Name)); errors.Is(err, os.ErrNotExist) {
		initConfig()
	}
}

func initConfig() {
	fmt.Println("Loading Config")
}

var upgrader = websocket.Upgrader{} // use default options

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	connectionUUID := uuid.New()
	connections[connectionUUID] = conn

	// The event loop
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Error("Error during message reading:", err)
			conn.Close()
			delete(connections, connectionUUID)
			break
		}
		log.Printf("Received: %s", message)
		for _, c := range connections {
			err = c.WriteMessage(messageType, message)
			if err != nil {
				log.Error("Error during message writing:", err)
				break
			}
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}
