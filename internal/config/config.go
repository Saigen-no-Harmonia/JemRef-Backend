package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPort     string
	DBPassword string
	DBName     string
	Env        string
	LogLevel   string
}

func Load() (Config, error) {
	// if os.Getenv("ENV") == "local" {
	// 	_ = godotenv.Load()
	// }

	_ = godotenv.Load()

	host := mustGetEnv("DB_HOST")
	user := mustGetEnv("DB_USER")
	port := mustGetEnv("DB_PORT")
	password := mustGetEnv("DB_PASS")
	name := mustGetEnv("DB_NAME")
	env := mustGetEnv("ENV")
	logLevel := mustGetEnv("LOG_LEVEL")

	return Config{
		DBHost:     host,
		DBUser:     user,
		DBPort:     port,
		DBPassword: password,
		DBName:     name,
		Env:        env,
		LogLevel:   logLevel,
	}, nil
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}
