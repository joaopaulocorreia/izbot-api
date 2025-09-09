package main

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Conn *amqp.Connection
}

type Message struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

func NewRabbitMQ(stringConnection string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(stringConnection)
	if err != nil {
		return &RabbitMQ{}, err
	}

	return &RabbitMQ{
		Conn: conn,
	}, nil
}

func (r RabbitMQ) Consume(queueName string) <-chan amqp.Delivery {
	ch, err := r.Conn.Channel()
	FailOnError(err, "[RabbitMQ][Consume] Fail to create queue")

	// queue, err := channel.QueueDeclarePassive(queueName, true, false, false, false, nil)
	// if err != nil {
	queue, err := ch.QueueDeclare(queueName, true, false, false, false, nil)
	FailOnError(err, "[RabbitMQ][Consume] Fail to create queue")
	// }

	dataChan, err := ch.Consume(queue.Name, "", true, false, false, false, nil)
	FailOnError(err, "[RabbitMQ][Consume] Fail to create consumer")

	return dataChan
}

func (r RabbitMQ) Publish(queueName string) {
	channel, err := r.Conn.Channel()
	FailOnError(err, "[RabbitMQ][Publish] Fail to create channel")

	err = channel.PublishWithContext(context.Background(), "", queueName, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte("testing hello"),
	})
	FailOnError(err, "[RabbitMQ][Publish] Failt to publish")
}
