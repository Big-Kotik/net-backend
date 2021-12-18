package msg

// ClientMessage struct for msg which came from client
type ClientMessage struct {
	Destination string `json:"destination"`
	Source      string `json:"source"`
	Message     string `json:"msg"`
}
