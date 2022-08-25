package common

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func (m *Message) JSONMarshal() ([]byte, error) {
	b := bytes.Buffer{}

	encoder := json.NewEncoder(&b)
	err := encoder.Encode(m)

	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func MessageFromJSON(data []byte) (*Message, error) {
	message := &Message{}
	if err := json.Unmarshal(data, message); err != nil {
		return nil, err
	}
	return message, nil
}

func RedisMarshalSenderReceiver(sender, receiver string) string {
	senderB64 := base64.StdEncoding.EncodeToString([]byte(sender))
	receiverB64 := base64.StdEncoding.EncodeToString([]byte(receiver))

	key := fmt.Sprintf("%s:%s", senderB64, receiverB64)

	return key
}

func RedisUnmarshalSenderReceiver(key string) (string, string, error) {
	sepIndex := strings.IndexRune(key, ':')
	if sepIndex == -1 {
		return "", "", errors.New("malformed key")
	}

	sender := key[:sepIndex]
	receiver := key[sepIndex+1:]

	return sender, receiver, nil
}

func (m *Message) RedisMarshal() (string, string) {
	key := RedisMarshalSenderReceiver(m.Sender, m.Receiver)

	return key, m.Message
}

func MessageFromRedis(key, value string) (*Message, error) {
	sender, receiver, err := RedisUnmarshalSenderReceiver(key)
	if err != nil {
		return nil, err
	}

	return &Message{
		Sender:   sender,
		Receiver: receiver,
		Message:  value,
	}, nil
}
