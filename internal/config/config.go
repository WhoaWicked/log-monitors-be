package config

import (
	"log"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	JWT   JWTConfig
}

type AppConfig struct {
	Port int `env:"APP_PORT,required"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST,required"`
	Port     int    `env:"DB_PORT,required"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME,required"`
}

type RedisConfig struct {
	Addr string `env:"REDIS_ADDR" envDefault:"localhost:6380"`
}

type JWTConfig struct {
	Secret     string `env:"JWT_SECRET,required"`
	ExpiryHour int    `env:"JWT_EXPIRY_HOUR" envDefault:"2"`
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}
	config := new(Config)
	if err := env.Parse(config); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return config
}
