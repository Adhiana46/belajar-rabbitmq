package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpConfig struct {
	host string
	port string
	user string
	pass string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	conf := amqpConfig{
		host: "localhost",
		port: "5672",
		user: "rabbit",
		pass: "firstrabbit",
	}

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", conf.user, conf.pass, conf.host, conf.port))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // auto-delete
		false,   // exclusive
		false,   // no-wait
		nil,     // args
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	time.Sleep(1 * time.Second)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for i := 0; i < 1000; i++ {
			body := "[0] Hello World " + strconv.Itoa(i)
			err = ch.PublishWithContext(ctx,
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				},
			)

			failOnError(err, "Failed to publish a message")
			log.Printf(" [x] Sent %s\n", body)
		}

		wg.Done()
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			body := "[1] Hello World " + strconv.Itoa(i)
			err = ch.PublishWithContext(ctx,
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "text/plain",
					Body:        []byte(body),
				},
			)

			failOnError(err, "Failed to publish a message")
			log.Printf(" [x] Sent %s\n", body)
		}

		wg.Done()
	}()

	wg.Wait()
}
