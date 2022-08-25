package common

import (
	"fmt"
	"github.com/streadway/amqp"
)

type MQState struct {
	connection   *amqp.Connection
	channel      *amqp.Channel
	messageQueue amqp.Queue
}

func NewMQState() (*MQState, error) {
	connection, err := amqp.Dial("amqp://user:password@localhost:7001/")
	if err != nil {
		return nil, err
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, err
	}

	messageQueue, err := channel.QueueDeclare("MessagesQueue",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, err
	}

	state := MQState{
		connection:   connection,
		channel:      channel,
		messageQueue: messageQueue,
	}

	return &state, nil
}

func (state *MQState) NewMessage(message Message) error {
	messageBytes, err := message.JSONMarshal()
	if err != nil {
		return err
	}

	fmt.Printf("The message being attempted to be sent is: %s", string(messageBytes))

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        messageBytes,
	}

	err = state.channel.Publish("", state.messageQueue.Name, false, false, msg)
	return err
}

func (state *MQState) EndState() {
	if err := state.channel.Close(); err != nil {
		_ = fmt.Errorf("failed to close RabbitMQ channel with error: %v", err)
	}

	if err := state.connection.Close(); err != nil {
		_ = fmt.Errorf("failed to close RabbitMQ connection with error: %v", err)
	}
}

func (state *MQState) Consume() (<-chan amqp.Delivery, error) {
	consumer, err := state.channel.Consume("", state.messageQueue.Name, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}
