package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DB                        DBConfig
	Port                      string
	JwtSecret                 []byte
	JwtAccessExpirationMinute int
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

func Load() (*Config, error) {
	cfg := &Config{
		DB: DBConfig{
			User:     getEnv("DB_USER", "user"),
			Password: getEnv("DB_PASSWORD", ""),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_POST", "5432"),
			Name:     getEnv("DB_NAME", "db"),
		},

		Port:                      getEnv("GO_PORT", "8080"),
		JwtSecret:                 []byte(getEnv("JWT_SECRET", "key")),
		JwtAccessExpirationMinute: getEnvAsInt("JWT_ACCESS_EXPIRATION_MINUTES", 15),
	}

	return cfg, nil
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.User, c.Password, c.Host, c.Port, c.Name,
	)
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}

	return defaultVal
}

func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}
