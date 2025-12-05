package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"rpc_assign/commons"
	"strings"
	"sync"
)

func main() {
	// Connect to server
	client, err := rpc.Dial("tcp", commons.Get_server_address())
	if err != nil {
		fmt.Printf("Error connecting to server: %v\n", err)
		return
	}
	defer client.Close()

	var name string
	fmt.Print("Enter your name: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	name = strings.TrimSpace(input)

	// Register with server
	var registerReply commons.RegisterReply
	err = client.Call("Chat.Register", commons.RegisterArgs{ClientName: name}, &registerReply)
	if err != nil {
		fmt.Printf("Error registering: %v\n", err)
		return
	}
	clientID := registerReply.ClientID
	fmt.Printf("Connected as %s\n\n", clientID)

	var wg sync.WaitGroup
	quitChan := make(chan bool)

	// Goroutine 1: Receive messages from server
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-quitChan:
				return
			default:
				var listenReply commons.ListenReply
				err := client.Call("Chat.Listen", clientID, &listenReply)
				if err != nil {
					return
				}

				msg := listenReply.Message
				switch msg.MessageType {
				case "join":
					fmt.Printf("\n[System] %s\n> ", msg.Text)
				case "leave":
					fmt.Printf("\n[System] %s\n> ", msg.Text)
				case "message":
					fmt.Printf("\n%s: %s\n> ", msg.ClientName, msg.Text)
				}
			}
		}
	}()

	// Goroutine 2: Send messages to server
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			fmt.Print("> ")
			input, _ := reader.ReadString('\n')
			text := strings.TrimSpace(input)

			if text == "quit" {
				// Unregister from server
				var unregisterReply bool
				client.Call("Chat.Unregister", clientID, &unregisterReply)
				close(quitChan)
				return
			}

			if text != "" {
				// Send message to server
				args := commons.SendMessageArgs{
					ClientID:   clientID,
					ClientName: name,
					Text:       text,
				}
				var sendReply bool
				client.Call("Chat.SendMessage", args, &sendReply)
			}
		}
	}()

	wg.Wait()
	fmt.Println("\nDisconnected from chat.")
}
