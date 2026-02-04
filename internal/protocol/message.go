package protocol

import "encoding/json"

// MessageType represents the type of LiveView message
type MessageType string

const (
	MessageTypeJoin      MessageType = "phx_join"
	MessageTypeReply     MessageType = "phx_reply"
	MessageTypeEvent     MessageType = "event"
	MessageTypeDiff      MessageType = "diff"
	MessageTypeHeartbeat MessageType = "heartbeat"
	MessageTypeClose     MessageType = "phx_close"
	MessageTypeError     MessageType = "phx_error"
	MessageTypeLeave     MessageType = "phx_leave"
)

// Message represents a LiveView protocol message
type Message struct {
	JoinRef *string         `json:"join_ref,omitempty"`
	Ref     *string         `json:"ref,omitempty"`
	Topic   string          `json:"topic"`
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

// Encode serializes a Message to JSON
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

// DecodeMessage deserializes JSON to a Message
func DecodeMessage(data []byte) (*Message, error) {
	var m Message
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// ReplyPayload represents a successful reply
type ReplyPayload struct {
	Status   string          `json:"status"`
	Response json.RawMessage `json:"response"`
}

// ErrorPayload represents an error reply
type ErrorPayload struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// EventPayload represents a client-sent event
type EventPayload struct {
	Type    string                 `json:"type"`
	Event   string                 `json:"event"`
	Value   map[string]interface{} `json:"value"`
	Target  string                 `json:"target,omitempty"`
	Targets []string               `json:"targets,omitempty"`
}

// DiffPayload represents a DOM diff update
type DiffPayload struct {
	Static  []interface{} `json:"s,omitempty"`
	Dynamic []interface{} `json:"d"`
}

// JoinPayload represents the initial join parameters
type JoinPayload struct {
	Params  map[string]interface{} `json:"params"`
	Session string                 `json:"session"`
	Static  string                 `json:"static"`
}

// NewJoinReply creates a successful join reply message
func NewJoinReply(topic string, ref string, rendered interface{}) *Message {
	response, _ := json.Marshal(map[string]interface{}{
		"rendered": rendered,
	})
	payload, _ := json.Marshal(ReplyPayload{
		Status:   "ok",
		Response: response,
	})
	return &Message{
		Ref:     &ref,
		Topic:   topic,
		Event:   "phx_reply",
		Payload: payload,
	}
}

// NewDiffMessage creates a diff update message
func NewDiffMessage(topic string, diff DiffPayload) (*Message, error) {
	payload, err := json.Marshal(diff)
	if err != nil {
		return nil, err
	}
	return &Message{
		Topic:   topic,
		Event:   "diff",
		Payload: payload,
	}, nil
}
