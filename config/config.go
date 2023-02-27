package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/pkg/errors"
)

const prefix = "MP_"

type Config struct {
	Telegram Telegram `envPrefix:"TELEGRAM_"`
	Database Database `envPrefix:"DB_"`
}

type Telegram struct {
	BotToken string `env:"BOT_TOKEN,notEmpty"`
	AdminID  int64  `env:"ADMIN_ID,notEmpty"`
}

type Database struct {
	Host     string `env:"HOST,notEmpty"`
	Port     int    `env:"PORT,notEmpty"`
	Database string `env:"DATABASE,notEmpty"`
	User     string `env:"USER,notEmpty"`
	Password string `env:"PASSWORD,notEmpty"`
}

func GetConfig() (Config, error) {
	config := Config{}

	err := env.Parse(&config, env.Options{
		Prefix: prefix,
	})

	return config, errors.Wrap(err, "env.Parse")
}
