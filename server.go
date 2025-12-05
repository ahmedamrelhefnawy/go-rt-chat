package main

import (
	"fmt"
	"net"
	"net/rpc"
	"rpc_assign/commons"
	"sync"
)

// ClientConnection represents a connected client
type ClientConnection struct {
	ID      string
	Name    string
	Channel chan commons.BroadcastMessage
}

// Chat is our RPC service
type Chat struct {
	clients map[string]*ClientConnection
	mutex   sync.Mutex
}

// NewChat creates a new Chat instance
func NewChat() *Chat {
	return &Chat{
		clients: make(map[string]*ClientConnection),
	}
}

// Register adds a new client to the chat
func (c *Chat) Register(args commons.RegisterArgs, reply *commons.RegisterReply) error {
	c.mutex.Lock()

	// Generate unique client ID
	clientID := fmt.Sprintf("%s", args.ClientName)

	// Create new client connection
	client := &ClientConnection{
		ID:      clientID,
		Name:    args.ClientName,
		Channel: make(chan commons.BroadcastMessage, 100),
	}

	c.clients[clientID] = client
	c.mutex.Unlock()

	reply.ClientID = clientID

	fmt.Printf("User %s joined\n", clientID)

	// Broadcast join message to all OTHER clients
	c.broadcastMessage(commons.BroadcastMessage{
		ClientID:    clientID,
		ClientName:  args.ClientName,
		Text:        fmt.Sprintf("User %s joined", clientID),
		MessageType: "join",
	}, clientID)

	return nil
}

// Unregister removes a client from the chat
func (c *Chat) Unregister(clientID string, reply *bool) error {
	c.mutex.Lock()
	client, exists := c.clients[clientID]
	if exists {
		close(client.Channel)
		delete(c.clients, clientID)
	}
	c.mutex.Unlock()

	if exists {
		fmt.Printf("User %s left\n", clientID)
		// Broadcast leave message
		c.broadcastMessage(commons.BroadcastMessage{
			ClientID:    clientID,
			ClientName:  client.Name,
			Text:        fmt.Sprintf("User %s left", clientID),
			MessageType: "leave",
		}, clientID)
	}

	*reply = true
	return nil
}

// Listen is a blocking call that waits for messages to broadcast to the client
func (c *Chat) Listen(clientID string, reply *commons.ListenReply) error {
	c.mutex.Lock()
	client, exists := c.clients[clientID]
	c.mutex.Unlock()

	if !exists {
		return fmt.Errorf("client not found")
	}

	// Block until a message is available
	msg, ok := <-client.Channel
	if !ok {
		return fmt.Errorf("client channel closed")
	}

	reply.Message = msg
	return nil
}

// SendMessage broadcasts a message to all other clients
func (c *Chat) SendMessage(args commons.SendMessageArgs, reply *bool) error {
	fmt.Printf("Message from %s: %s\n", args.ClientName, args.Text)

	// Broadcast to all OTHER clients (no self-echo)
	c.broadcastMessage(commons.BroadcastMessage{
		ClientID:    args.ClientID,
		ClientName:  args.ClientName,
		Text:        args.Text,
		MessageType: "message",
	}, args.ClientID)

	*reply = true
	return nil
}

// broadcastMessage sends a message to all clients except the sender
func (c *Chat) broadcastMessage(msg commons.BroadcastMessage, excludeClientID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for id, client := range c.clients {
		if id != excludeClientID {
			// Non-blocking send to avoid deadlock if client is slow
			select {
			case client.Channel <- msg:
			default:
				fmt.Printf("Warning: Could not send to client %s (channel full)\n", id)
			}
		}
	}
}

func main() {
	addr := commons.Get_server_address()
	listener, _ := net.Listen("tcp", addr)
	fmt.Printf("Server running on %s...\n", addr)

	chat := NewChat()
	rpc.Register(chat)

	for {
		conn, _ := listener.Accept()
		go rpc.ServeConn(conn)
	}
}
