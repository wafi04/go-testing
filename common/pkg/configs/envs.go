package configs

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
    DatabaseURL   string
    RedisURL      string
    JWTSecret     string
    ServerPort    string
}

func LoadEnv(serviceName string) *EnvConfig {
    envPath := fmt.Sprintf("./%s/.env", serviceName)
    if err := godotenv.Load(envPath); err != nil {
        log.Printf("No .env file found for %s: %v", serviceName, err)
    }

    return &EnvConfig{
        DatabaseURL:  os.Getenv("DATABASE_URL"),
        RedisURL:     os.Getenv("REDIS_URL"),
        JWTSecret:    os.Getenv("JWT_SECRET"),
        ServerPort:   os.Getenv("SERVER_PORT"),
    }
}

