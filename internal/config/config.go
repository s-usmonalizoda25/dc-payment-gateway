package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env:"ENV" env-default:"development"`
	ServerPort string `env:"SERVER_PORT" env-default:":8080"`
	DB         DBConfig
	ExpressPay ExpressPayConfig
}

type ExpressPayConfig struct {
	Login    string `env:"EXPRESSPAY_LOGIN"`
	Password string `env:"EXPRESSPAY_PASSWORD"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST" env-default:"localhost"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	User     string `env:"DB_USER" env-default:"postgres"`
	Password string `env:"DB_PASSWORD" env-default:"postgres"`
	Name     string `env:"DB_NAME" env-default:"dc_payment_gateway"`
	SSLMode  string `env:"DB_SSLMODE" env-default:"disable"`
}

func (db DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode)
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
		if instance.ExpressPay.Login == "" || instance.ExpressPay.Password == "" {
			log.Fatal("EXPRESSPAY_LOGIN and EXPRESSPAY_PASSWORD must be set in .env or environment variables")
		}
	})
	return instance
}
