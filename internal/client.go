package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	//tea "github.com/charmbracelet/bubbletea"
)

type Client struct {
	status   bool
	message  string
	username string
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
	conf, _ := GetConfig()
	done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	socketUrl := conf.CLIENT_HOST_PORT + "/socket"
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

// CLient

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getDimensions() (uint, uint) {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	return uint(ws.Row), uint(ws.Col)
}

func ClientStart(username string) {
	conf, _ := GetConfig()
	client := Client{status: true, message: "Open", username: username}
	fmt.Println(client.status, client.message)
	ch := make(chan string)
	wg := new(sync.WaitGroup)
	socketChannel := make(chan string)
	mainChannel := make(chan string)
	socketUrl := conf.CLIENT_HOST_PORT + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	wg.Add(1)
	//go dummySocket(socketChannel, mainChannel, wg)
	go socket(socketChannel, mainChannel, conn, wg)
	wg.Add(3)
	go recieveLoop(ch, socketChannel, wg)
	go sendLoop(ch, mainChannel, conn, wg, username)
	go programLoop(ch, wg)
	wg.Wait()
	fmt.Println("Exiting Skyline")
}

func recieveLoop(ch chan string, socketChannel chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		res, ok := <-socketChannel
		if !ok {
			fmt.Println("Channel Close ", ok)
			break
		}
		ch <- res
	}
	close(ch)
}

func sendLoop(ch chan string, mainChannel chan string, conn *websocket.Conn, wg *sync.WaitGroup, username string) {
	defer wg.Done()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == ":quit" {
			fmt.Println("Quitting...")
			close(mainChannel)
			break
		}
		fmt.Printf("\033[F\033[K")
		err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("[%s]: %q\n", username, line))) //systemUser
		if err != nil {
			log.Println("Error during writing to websocket:", err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error encountered:", err)
	}
}

func programLoop(ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		_, ok := <-ch
		if !ok {
			fmt.Println("Channel Close ", ok)
			break
		}
	}
}

func socket(socketChannel chan string, mainChannel chan string, conn *websocket.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
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
	}(conn)
	for {
		_, ok := <-mainChannel
		if !ok {
			close(socketChannel)
			break
		}
	}
}
