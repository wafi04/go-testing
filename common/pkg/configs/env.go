// common/pkg/configs/env.go
package configs

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/wafi04/go-testing/common/pkg/logger"
)

func GetEnv(key string, servicePath string) string {
	log := logger.NewLogger()

	envPath := filepath.Join(servicePath, ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Log(logger.ErrorLevel, "Failed to load env from %s: %v", envPath, err)
	}

	return os.Getenv(key)
}
