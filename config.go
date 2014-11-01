package main

import (
	"log"

	"github.com/Unknwon/goconfig"
)

type Config struct {
	TestDataPath string
	TempPath     string
	RunPath      string
	RMQHost      string
	RMQPort      int
	RMQUser      string
	RMQPassword  string
	RMQQueueName string
	SQLHost      string
	SQLPort      int
	SQLUser      string
	SQLPassword  string
	SQLDatabase  string
	RedisHost    string
	RedisPort    int
}

func ParseConfig(path string) (config *Config, err error) {
	c, err := goconfig.LoadConfigFile(path)
	config = &Config{}
	config.TestDataPath = c.MustValue("default", "testdata", "./testdata/")
	config.TempPath = c.MustValue("default", "tmp", "./tmp/")
	config.RunPath = c.MustValue("default", "runpath", "./rundir/")

	config.RMQHost = c.MustValue("rabbitmq", "host", "localhost")
	config.RMQPort = c.MustInt("rabbitmq", "port", 5672)
	config.RMQUser = c.MustValue("rabbitmq", "user", "guest")
	config.RMQPassword = c.MustValue("rabbitmq", "password", "guest")
	config.RMQQueueName = c.MustValue("rabbitmq", "queue", "judge_task")

	config.SQLHost = c.MustValue("mysql", "host", "localhost")
	config.SQLPort = c.MustInt("mysql", "port", 3306)
	config.SQLUser = c.MustValue("mysql", "user")
	if config.SQLUser == "" {
		log.Fatalln("Config Error: No mysql user")
	}
	config.SQLPassword = c.MustValue("mysql", "password")
	if config.SQLPassword == "" {
		log.Fatalln("Config Error: No mysql password")
	}
	config.SQLDatabase = c.MustValue("mysql", "database")
	if config.SQLDatabase == "" {
		log.Fatalln("Config Error: No mysql database")
	}

	config.RedisHost = c.MustValue("redis", "host", "localhost")
	config.RedisPort = c.MustInt("redis", "posr", 6379)

	return
}
