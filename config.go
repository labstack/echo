package echo

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config interface {
	// GetEnv returns the value of given environment vairable
	GetEnv(key string) string

	// GetEnvOrDefault returns the value of given environment vairable, returns the val otherwise
	GetEnvOrDefault(key, val string) string
}

type configProvider struct {
}

func DefaultConfigProvider(logger Logger) Config {
	cp := new(configProvider)

	read("./configs/.env", logger)

	env := os.Getenv("APP_ENV")
	if env == "" {
		return cp
	}

	read("./configs/."+env+".env", logger)

	return cp
}

func read(file string, logger Logger) {
	err := godotenv.Load(file)
	if err != nil {
		log.Fatalf("Error loading .env file. [%v]", err)
	}
}

func (c *configProvider) GetEnv(key string) string {
	return os.Getenv(key)
}

func (c *configProvider) GetEnvOrDefault(key, val string) string {
	v := c.GetEnv(key)
	if v == "" {
		return val
	}

	return v
}
