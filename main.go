package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"time"

	"database/sql"

	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

const APP_VER = "0.4.0"

var (
	uid    int
	gid    int
	db     *sql.DB
	rdb    redis.Conn
	pool   *redis.Pool
	config *Config
)

func init() {
	usr, _ := user.Lookup(os.Getenv("SUDO_USER"))
	uid, _ = strconv.Atoi(usr.Uid)
	gid, _ = strconv.Atoi(usr.Gid)
}

func initDir() {
	if _, err := os.Stat(config.TempPath); err != nil && !os.IsExist(err) {
		os.MkdirAll(config.TempPath, os.ModePerm)
	}
	os.Chown(config.TempPath, uid, gid)
	if _, err := os.Stat(config.RunPath); err != nil && !os.IsExist(err) {
		os.MkdirAll(config.RunPath, os.ModePerm)
	}
	os.Chown(config.RunPath, uid, gid)
}

func newPool(host string, port int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     2,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			cs := fmt.Sprintf("%s:%d", host, port)
			c, err := redis.Dial("tcp", cs)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "JudgeNode"
	app.Usage = "A distribute online judge node"
	app.Version = APP_VER
	app.Author = "maemual (maemual@gmail.com)"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "c",
			Value: "conf.ini",
			Usage: "The path of config file",
		},
	}
	app.Action = func(c *cli.Context) {
		var err error
		config, err = ParseConfig(c.String("c"))
		initDir()

		// MySQL
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			config.SQLUser,
			config.SQLPassword,
			config.SQLHost,
			config.SQLPort,
			config.SQLDatabase)
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		err = db.Ping()
		if err != nil {
			log.Fatal(err)
		}

		// Redis connection pool
		pool = newPool(config.RedisHost, config.RedisPort)

		// RabbitMQ
		url := fmt.Sprintf("amqp://%s:%s@%s:%d/",
			config.RMQUser,
			config.RMQPassword,
			config.RMQHost,
			config.RMQPort)
		conn, err := amqp.Dial(url)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		ch, err := conn.Channel()
		failOnError(err, "Failed to open a channel")
		defer ch.Close()

		q, err := ch.QueueDeclare(
			config.RMQQueueName, // name
			true,                // durable
			false,               // delete when unused
			false,               // exclusive
			false,               // no-wait
			nil,                 // arguments
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
				dealMessage(d.Body)
				d.Ack(false)
			}
		}()

		log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
		<-forever
	}

	app.Run(os.Args)
}
