package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config описывает общую конфигурацию для API-сервиса.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Valkey   ValkeyConfig   `mapstructure:"valkey"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type AppConfig struct {
	Env      string        `mapstructure:"env"`
	Port     int           `mapstructure:"port"`
	LogLevel string        `mapstructure:"log_level"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	DB       string `mapstructure:"db"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DB, p.SSLMode,
	)
}

type ValkeyConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

func (v ValkeyConfig) Addr() string {
	return fmt.Sprintf("%s:%d", v.Host, v.Port)
}

type MinIOConfig struct {
	Endpoint     string `mapstructure:"endpoint"`
	RootUser     string `mapstructure:"root_user"`
	RootPassword string `mapstructure:"root_password"`
	Bucket       string `mapstructure:"bucket"`
	UseSSL       bool   `mapstructure:"use_ssl"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	AccessTTL  time.Duration `mapstructure:"access_ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

// Load загружает конфигурацию из файла (если есть) и переменных окружения.
func Load() (*Config, error) {
	v := viper.New()

	// Значения по умолчанию
	v.SetDefault("app.env", "development")
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.log_level", "debug")
	v.SetDefault("app.timeout", 15*time.Second)

	v.SetDefault("postgres.host", "localhost")
	v.SetDefault("postgres.port", 5432)
	v.SetDefault("postgres.db", "school")
	v.SetDefault("postgres.user", "school")
	v.SetDefault("postgres.password", "school_dev_password")
	v.SetDefault("postgres.sslmode", "disable")

	v.SetDefault("valkey.host", "localhost")
	v.SetDefault("valkey.port", 6379)
	v.SetDefault("valkey.password", "")

	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.root_user", "minioadmin")
	v.SetDefault("minio.root_password", "minioadmin")
	v.SetDefault("minio.bucket", "school-files")
	v.SetDefault("minio.use_ssl", false)

	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.access_ttl", 15*time.Minute)
	v.SetDefault("jwt.refresh_ttl", 720*time.Hour)

	// Чтение переменных окружения (например, APP_PORT, POSTGRES_HOST, VALKEY_HOST)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Опционально читаем .env в корне или в рабочей директории
	v.AddConfigPath(".")
	v.AddConfigPath("../")
	v.SetConfigName(".env")
	v.SetConfigType("env")
	_ = v.ReadInConfig() // Ошибку отсутствия файла игнорируем, переменные могут быть в env

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
