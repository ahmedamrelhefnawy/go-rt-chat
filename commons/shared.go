package commons

/*
TODO
you may define any structs here to be used by the rpc
*/

// Message represents a chat message
type Message struct {
	ClientName string
	Text       string
}

// SendMessageArgs contains arguments for sending a message
type SendMessageArgs struct {
	ClientID   string
	ClientName string
	Text       string
}

// GetMessagesArgs contains arguments for fetching messages
type GetMessagesArgs struct {
	StartIndex int // Starting index for fetching messages (0 for all)
	Count      int // Number of messages to fetch (0 or negative for all)
}

// RegisterArgs contains arguments for registering a client
type RegisterArgs struct {
	ClientName string
}

// RegisterReply contains the response for registration
type RegisterReply struct {
	ClientID string
}

// BroadcastMessage represents a message to be broadcast to clients
type BroadcastMessage struct {
	ClientID    string
	ClientName  string
	Text        string
	MessageType string // "join", "message", "leave"
}

// ListenReply contains the broadcast message
type ListenReply struct {
	Message BroadcastMessage
}

// hint: you will need to have the server address fixed between clients and the coordinating server
func Get_server_address() string {
	return "0.0.0.0:9999"
}
