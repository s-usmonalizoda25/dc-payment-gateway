package config

import (
	"fmt"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"development"`
	ServerPort string `yaml:"server_port" env:"SERVER_PORT" env-default:":8080"`
	DB         DatabaseConfig
	ExpressPay ExpressPayConfig
}

type ExpressPayConfig struct {
	Login    string `env:"EXPRESSPAY_LOGIN" env-default:"21950"`
	Password string `env:"EXPRESSPAY_PASSWORD" env-default:"147521"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST" env-default:"localhost"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	User     string `env:"DB_USER" env-default:"postgres"`
	Password string `env:"DB_PASSWORD" env-default:"postgres"`
	Name     string `env:"DB_NAME" env-default:"dc_payment_db"`
	SSLMode  string `env:"DB_SSLMODE" env-default:"disable"`
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig(".env", instance); err != nil {
			_ = cleanenv.ReadEnv(instance)
		}
	})
	return instance
}
	