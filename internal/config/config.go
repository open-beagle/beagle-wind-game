package config

import (
	"os"
)

// Config 存储应用程序配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port    string
	Host    string
	BaseURL string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey string
	ExpiresIn int // token过期时间（小时）
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:    getEnvOrDefault("SERVER_PORT", "8080"),
			Host:    getEnvOrDefault("SERVER_HOST", "localhost"),
			BaseURL: getEnvOrDefault("BASE_URL", "http://localhost:8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			DBName:   getEnvOrDefault("DB_NAME", "beagle_wind_game"),
		},
		JWT: JWTConfig{
			SecretKey: getEnvOrDefault("JWT_SECRET", "your-secret-key"),
			ExpiresIn: 24, // 默认24小时
		},
	}

	return config, nil
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
