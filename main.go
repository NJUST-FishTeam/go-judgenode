package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

const APP_VER = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Name = "JudgeNode"
	app.Usage = "A distribute online judge node"
	app.Version = APP_VER
	app.Author = "maemual (maemual@gmail.com)"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "queue",
			Value: "judge_task",
			Usage: "The name of judge task queue",
		},
		cli.StringFlag{
			Name:  "host",
			Value: "localhost",
			Usage: "the ip of the RabbitMQ",
		},
		cli.StringFlag{
			Name:  "port",
			Value: "5672",
			Usage: "the port of the RabbitMQ",
		},
		cli.StringFlag{
			Name:  "user",
			Value: "guest",
			Usage: "the user name of the RabbitMQ",
		},
		cli.StringFlag{
			Name:  "password",
			Value: "guest",
			Usage: "the password of the RabbitMQ",
		},
		cli.StringFlag{
			Name:  "datapath",
			Value: "./testdata/",
			Usage: "The path of test data",
		},
	}
	app.Action = func(c *cli.Context) {
		conn, err := amqp.Dial("amqp://" +
			c.String("user") + ":" +
			c.String("password") + "@" +
			c.String("host") + ":" +
			c.String("port") + "/")
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		ch, err := conn.Channel()
		failOnError(err, "Failed to open a channel")
		defer ch.Close()

		q, err := ch.QueueDeclare(
			c.String("queue"), // name
			true,              // durable
			false,             // delete when unused
			false,             // exclusive
			false,             // no-wait
			nil,               // arguments
		)
		failOnError(err, "Failed to declare a queue")

		err = ch.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
		)
		failOnError(err, "Failed to set QoS")

		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		failOnError(err, "Failed to register a consumer")

		forever := make(chan bool)

		go func() {
			for d := range msgs {
				log.Printf("Received a message: %s", d.Body)
				dealMessage(d.Body, c.String("datapath"))
				d.Ack(false)
			}
		}()

		log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
		<-forever
	}

	app.Run(os.Args)
}
