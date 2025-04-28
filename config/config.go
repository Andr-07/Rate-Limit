package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Redis RedisConfig
}

type RedisConfig struct {
	Addr     string
	Password string
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
	}
}
