package config

import "os"

type RabbitMQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string

	//交换机
	TaggingExchange string
	IndexExchange   string

	//队列
	TaggingQueue string
	IndexQueue   string

	//路由键
	TaggingKey string
	IndexKey   string
}

var Config = RabbitMQConfig{
	Host:     getEnv("RABBITMQ_HOST", "localhost"),
	Port:     5672,
	User:     "guest",
	Password: "guest",
	VHost:    "/",

	TaggingExchange: "tagging_exchange",
	IndexExchange:   "index_exchange",

	TaggingQueue: "tagging_queue",
	IndexQueue:   "index_queue",

	TaggingKey: "tagging.task",
	IndexKey:   "index.update",
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
