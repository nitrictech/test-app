package common

type Message struct {
	MessageType string `json:"messageType"`
	ID          string `json:"id"`
	PayloadType string `json:"payloadType"`
	Payload     string `json:"payload"`
}
