package client

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"skyline/config"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	status        bool
	message       string
	username      string
	ch            chan string
	wg            *sync.WaitGroup
	socketChannel chan string
	mainChannel   chan string
	socketUrl     string
	conn          *websocket.Conn
}

var done chan interface{}
var interrupt chan os.Signal

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}

func clientSocket() {
	conf, _ := config.GetConfig()
	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	socketUrl := conf.SERVER_URL
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()
	go receiveHandler(conn)

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case <-time.After(time.Duration(1) * time.Millisecond * 1000):
			// Send an echo packet every second
			err := conn.WriteMessage(websocket.TextMessage, []byte("Hello from GolangDocs!"))
			if err != nil {
				log.Println("Error during writing to websocket:", err)
				return
			}

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return
		}
	}
}

func ClientStart(username string) {
	conf, _ := config.GetConfig()
	client := Client{status: true, message: "Open", username: username}
	fmt.Println(client.status, client.message)
	client.ch = make(chan string)
	client.wg = new(sync.WaitGroup)
	client.socketChannel = make(chan string)
	client.mainChannel = make(chan string)
	client.socketUrl = conf.SERVER_URL
	var err error
	client.conn, _, err = websocket.DefaultDialer.Dial(client.socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	client.wg.Add(1)
	//go dummySocket(socketChannel, mainChannel, wg)
	go client.socket()
	client.wg.Add(3)
	go client.recieveLoop()
	go client.sendLoop()
	go client.programLoop()
	client.wg.Wait()
	fmt.Println("Exiting Skyline")
}

func (c *Client) recieveLoop() {
	defer c.wg.Done()
	for {
		res, ok := <-c.socketChannel
		if !ok {
			fmt.Println("Channel Close ", ok)
			break
		}
		c.ch <- res
	}
	close(c.ch)
}

func (c *Client) sendLoop() {
	defer c.wg.Done()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == ":quit" {
			fmt.Println("Quitting...")
			close(c.mainChannel)
			break
		}
		fmt.Printf("\033[F\033[K")
		err := c.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("[%s]: %q\n", c.username, line))) //systemUser
		if err != nil {
			log.Println("Error during writing to websocket:", err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error encountered:", err)
	}
}

func (c *Client) programLoop() {
	defer c.wg.Done()
	for {
		_, ok := <-c.ch
		if !ok {
			fmt.Println("Channel Close ", ok)
			break
		}
	}
}

func (c *Client) socket() {
	defer c.wg.Done()
	go func(conn *websocket.Conn) {
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error in receive:", err)
				return
			}
			log.Printf("Received: %s", msg)
		}
	}(c.conn)
	for {
		_, ok := <-c.mainChannel
		if !ok {
			close(c.socketChannel)
			break
		}
	}
}

func GenerateUserID() *string {
	userID := uuid.New().String()
	return &userID
}
