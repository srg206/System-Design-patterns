package kafka

import "fmt"

type Message struct {
	Topic   string
	Key     *string
	Value   []byte
	Headers map[string][]byte
}

func (m *Message) Validate() error {
	if m.Topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	if m.Value == nil {
		return fmt.Errorf("value cannot be nil")
	}
	return nil
}

func NewMessage(topic string, key *string, value []byte, headers map[string][]byte) *Message {
	return &Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: headers,
	}
}
