package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type PostgresConfig struct {
	User     string
	Password string
	Port     string
	Host     string
	DBName   string
}

type TimeoutsConfig struct {
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	ContextTimeout time.Duration
}

var (
	Drssr                ServerConfig
	Redis                RedisConfig
	Postgres             PostgresConfig
	ExpirationCookieTime time.Duration
	Timeouts             TimeoutsConfig
)

func SetConfig() {
	viper.SetConfigFile("config.json")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	Drssr = ServerConfig{
		Port: viper.GetString(`drssr.port`),
	}

	Redis = RedisConfig{
		Addr:     viper.GetString(`redis.address`),
		Password: viper.GetString(`redis.password`),
		DB:       viper.GetInt(`redis.db_name`),
	}

	Postgres = PostgresConfig{
		Port:     viper.GetString(`postgres.port`),
		Host:     viper.GetString(`postgres.host`),
		User:     viper.GetString(`postgres.user`),
		Password: viper.GetString(`postgres.pass`),
		DBName:   viper.GetString(`postgres.name`),
	}

	ExpirationCookieTime = viper.GetDuration("expiration_cookie_time")

	Timeouts = TimeoutsConfig{
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    5 * time.Second,
		ContextTimeout: time.Second * 2,
	}
}
