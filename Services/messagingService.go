package Services

import (
	"encoding/json"
	"fmt"
	"log"

	"../Models"
	"github.com/streadway/amqp"
)

var conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
var ch, chErr = conn.Channel()

func GetConnection() *amqp.Connection {
	return conn
}

func GetChannel() *amqp.Channel {
	return ch
}

func Publish(message string, channel string, event string) {
	q, err := ch.QueueDeclare(
		channel,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")
	data := &Models.Message{Message: message, Event: "msg"}
	b, err := json.Marshal(data)

	body := string(b)
	ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}

func Subscribe(channel string) <-chan amqp.Delivery {
	q, err := ch.QueueDeclare(
		channel,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	return msgs
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
