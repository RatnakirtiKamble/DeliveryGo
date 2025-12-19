package app

import (
	"log"
	"os"
	"strings"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr    	string
	PostgresDSN 	string
	RedisAddr		string 
	OSRMAddr 		string
	KafkaBrokers	[]string
}

func LoadConfig() Config {
	_ = godotenv.Load()

	cfg := Config{
		HTTPAddr:    getEnv("HTTP_ADDR", ":8000"),
		PostgresDSN: getEnv("POSTGRES_DSN", ""),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		OSRMAddr:    getEnv("OSRM_ADDR", ""),
		KafkaBrokers: strings.Split(
			getEnv("KAFKA_BROKERS", "localhost:9092"),
			",",
		),
	}

	if cfg.PostgresDSN == "" {
		log.Fatal("POSTGRES_DSN is required")
	}

	if cfg.OSRMAddr == "" {
		log.Fatal("OSMRAddr is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
