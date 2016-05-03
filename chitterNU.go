package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var idChan = make(chan string)

func IdManager() {
	var i int
	for i = 0; ; i++ {
		idChan <- strconv.Itoa(i)
	}
}

type Message struct {
	cmd string
	msg string
	dst string
	src string
}

type Client struct {
	ch   chan Message
	id   string
	conn net.Conn
}

func HandleAll(register_channel chan Client, rec chan Message) {
	hash := make(map[string]chan Message) //save the channel for each client using map
	for {
		select {
		case new_client := <-register_channel:
			hash[new_client.id] = new_client.ch
			fmt.Println("New Client connected", new_client.id)
		case message := <-rec:
			if strings.Compare(message.cmd, "0") == 0 {
				message.msg = message.src + ":" + message.msg
				for _, value := range hash {
					value <- message
				}
			} else if strings.Compare(message.cmd, "1") == 0 {
				message.msg = "chitter:" + message.src + "\n"
				hash[message.src] <- message
			} else if strings.Compare(message.cmd, "2") == 0 {
				arr := strings.SplitN(message.msg, ":", 2)
				message.msg = message.src + ":" + arr[1]
				for _, value := range hash {
					value <- message
				}
			} else if strings.Compare(message.cmd, "3") == 0 {
				arr := strings.SplitN(message.msg, ":", 2)
				message.msg = message.src + ":" + arr[1]
				if _, ok := hash[message.dst]; ok {
					hash[message.dst] <- message
				} else {
					message.msg = "Error! No such client!\n"
					hash[message.src] <- message
				}
			} else if strings.Compare(message.cmd, "5") == 0 {
				delete(hash, message.src)
			} else {
				message.msg = "Error! No such command.\n"
				hash[message.src] <- message
			}
		}
	}
}

//handle new client
func NewClient(rec chan Message, conn net.Conn, send chan Message, id string) {
	b := bufio.NewReader(conn)
	go HandleRec(rec, conn)
	for {
		line, err := b.ReadBytes('\n')
		if err != nil {
			conn.Close()
			fmt.Println("Connection closed:", id)
			close_flag := Message{cmd: "5", msg: "Disconnect!", dst: id, src: id}
			send <- close_flag
			break
		}
		input := string(line)
		message := Message{cmd: "0", msg: input, dst: "!", src: id}
		//logic for command
		index := strings.Index(input, ":")
		if index != -1 {
			arr := strings.SplitN(input, ":", 2)
			command := strings.TrimSpace(arr[0])
			if strings.HasPrefix(input, "whoami:") {
				message.cmd = "1"
			} else if strings.Compare(command, "all") == 0 {
				message.cmd = "2"
			} else if _, err := strconv.Atoi(command); err == nil {
				message.cmd = "3"
				message.dst = command
			} else {
				message.cmd = "4"
			}
		}
		send <- message
	}
}

//handle received message
func HandleRec(rec chan Message, conn net.Conn) {
	for {
		words := <-rec
		conn.Write([]byte(words.msg))
	}
}

func main() {
	rec := make(chan Message)
	register_channel := make(chan Client) //channel for letting the HandleAll know that there is a new client
	if len(os.Args) < 2 {
		fmt.Println("Please enter port number !")
		os.Exit(1)
	}
	port := os.Args[1]
	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Cannot connect to port", port)
		os.Exit(1)
	}
	go IdManager()
	fmt.Println("Using port", port)
	go HandleAll(register_channel, rec)
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Error in connection")
			os.Exit(1)
		} else {
			fmt.Println("Accepted new connection")
		}
		id := <-idChan
		new_client := Client{ch: make(chan Message), id: id, conn: conn}
		register_channel <- new_client
		go NewClient(new_client.ch, conn, rec, id) //create new thread for new client
	}
}
