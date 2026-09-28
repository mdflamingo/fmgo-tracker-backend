package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Postgres struct {
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
}

type Keycloak struct {
	KeycloakURL string
	ClientID    string
}

type Minio struct {
	MinioEndpoint     string
	MinioRootUser     string
	MinioRootPassword string
	MinioBucket       string
	// MinioUseSSL       bool
}

type Config struct {
	RunAddr    string
	LogLevel   string
	Debug      string
	ServerAddr string

	Postgres Postgres
	Minio    Minio
	Keycloak Keycloak
}

func GetConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Missing required .env file: ", err)

	}
	pgConfig := Postgres{
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
	}
	minioConfig := Minio{
		MinioEndpoint:     os.Getenv("MINIO_ENDPOINT"),
		MinioRootUser:     os.Getenv("MINIO_ROOT_USER"),
		MinioRootPassword: os.Getenv("MINIO_ROOT_PASSWORD"),
		MinioBucket:       os.Getenv("MINIO_BUCKET"),
		// MinioUseSSL:       os.Getenv("MINIO_SSL"),
	}
	keycloakConfig := Keycloak{
		KeycloakURL: os.Getenv("KEYCLOAK_URL"),
		ClientID:    os.Getenv("CLIENT_ID"),
	}
	mainConfig := Config{
		RunAddr:    os.Getenv("RUN_ADDR"),
		LogLevel:   os.Getenv("LOG_LEVEL"),
		ServerAddr: os.Getenv("SERVER_ADDR"),
		Debug:      os.Getenv("DEBUG"),
		Minio:      minioConfig,
		Postgres:   pgConfig,
		Keycloak:   keycloakConfig,
	}

	return &mainConfig
}
