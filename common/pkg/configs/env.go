package configs

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/wafi04/go-testing/common/pkg/logger"
)

func GetEnv(key string) string {
	log := logger.NewLogger()
	err := godotenv.Load(".env")
	if err != nil {
		log.Log(logger.ErrorLevel, "Failed to load env : %v", err)
	}
	return os.Getenv(key)
}
