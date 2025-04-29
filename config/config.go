package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Redis RedisConfig
	Kafka KafkaConfig
}

type RedisConfig struct {
	Addr     string
	Password string
}

type KafkaConfig struct {
	Addr string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	return &Config{
		Redis: RedisConfig{
			Addr:     os.Getenv("REDIS_ADDR"),
			Password: os.Getenv("REDIS_PASS"),
		},
		Kafka: KafkaConfig{
			Addr: os.Getenv("KAFKA_ADDR"),
		},
	}
}
