package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"sync"

	"skyline/config"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	connections = make(map[uuid.UUID]*websocket.Conn)
	connMutex   sync.Mutex
	upgrader    = websocket.Upgrader{} // use default options
)

func ServerInit() {
	conf, err := config.GetConfig()
	if err != nil {
		log.Fatal("Error loading config: ", err)
	}
	ConfigInit()
	log.Debugln("debug")
	connections = make(map[uuid.UUID]*websocket.Conn)
	http.HandleFunc("/socket", socketHandler)
	http.HandleFunc("/", home)
	log.Debugln("Server started at", conf.SERVER_HOST_PORT)
	err = http.ListenAndServeTLS(conf.SERVER_HOST_PORT, "cert.pem", "key.pem", nil)
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
	// TODO: Implement configuration file creation
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade our raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error during connection upgradation:", err)
		return
	}
	defer conn.Close()

	connectionUUID := uuid.New()
	connMutex.Lock()
	connections[connectionUUID] = conn
	connMutex.Unlock()
	log.Debugf("New WebSocket connection established with UUID: %s", connectionUUID)

	var wg sync.WaitGroup
	messageChan := make(chan []byte)

	wg.Add(2)
	go readMessages(conn, messageChan, &wg)
	go broadcastMessages(messageChan, connectionUUID, &wg)

	wg.Wait()
	connMutex.Lock()
	delete(connections, connectionUUID)
	connMutex.Unlock()
	conn.Close()
}

func readMessages(conn *websocket.Conn, messageChan chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(messageChan)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Error("Error during message reading:", err)
			return
		}
		messageChan <- message
	}
}

func broadcastMessages(messageChan chan []byte, connectionUUID uuid.UUID, wg *sync.WaitGroup) {
	defer wg.Done()
	for message := range messageChan {
		log.Printf("Received: %s", message)
		connMutex.Lock()
		for uuid, c := range connections {
			if uuid != connectionUUID {
				go func(c *websocket.Conn) {
					err := c.WriteMessage(websocket.TextMessage, message)
					if err != nil {
						log.Error("Error during message writing:", err)
					}
				}(c)
			}
		}
		connMutex.Unlock()
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Index Page")
}
