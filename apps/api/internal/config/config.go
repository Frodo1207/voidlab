package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	Environment   string
	DatabasePath  string
	UploadsDir    string
	PublicBaseURL string
	MinioEndpoint string
	MinioBucket   string
}

func Load() Config {
	port := getEnv("PORT", "18081")

	return Config{
		Port:          port,
		Environment:   getEnv("APP_ENV", "development"),
		DatabasePath:  getEnv("DB_PATH", "data/sqlite/voidlab.db"),
		UploadsDir:    getEnv("UPLOADS_DIR", "data/uploads"),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", fmt.Sprintf("http://localhost:%s", port)),
		MinioEndpoint: getEnv("MINIO_ENDPOINT", "http://minio:9000"),
		MinioBucket:   getEnv("MINIO_BUCKET", "voidlab-media"),
	}
}

func (c Config) Address() string {
	return fmt.Sprintf(":%s", c.Port)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
