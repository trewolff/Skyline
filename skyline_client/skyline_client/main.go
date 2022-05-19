package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"

	//tea "github.com/charmbracelet/bubbletea"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	cliFunc()
	initFunc()
	start()
}

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

func start() {
	client := Client{status: true, message: "Open"}
	fmt.Println(client.status, client.message)
	ch := make(chan string)
	wg := new(sync.WaitGroup)
	socketChannel := make(chan string)
	mainChannel := make(chan string)
	socketUrl := "ws://localhost:8080" + "/socket"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	wg.Add(1)
	//go dummySocket(socketChannel, mainChannel, wg)
	go socket(socketChannel, mainChannel, conn, wg)
	wg.Add(3)
	go recieveLoop(ch, socketChannel, wg)
	go sendLoop(ch, mainChannel, conn, wg)
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

func sendLoop(ch chan string, mainChannel chan string, conn *websocket.Conn, wg *sync.WaitGroup) {
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
		err := conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("[%s]: %q\n", systemUser, line)))
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
