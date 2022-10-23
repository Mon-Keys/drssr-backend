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

type CutterConfig struct {
	URL     string
	Timeout time.Duration
}

type ClassifierConfig struct {
	URL string
}

type SimilarityConfig struct {
	URL     string
	Timeout time.Duration
}

type TimeoutsConfig struct {
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	ContextTimeout time.Duration
}

var (
	Drssr                 ServerConfig
	Redis                 RedisConfig
	Postgres              PostgresConfig
	Cutter                CutterConfig
	Classifier            ClassifierConfig
	Similarity            SimilarityConfig
	WellSimilarityPercent int
	ExpirationCookieTime  time.Duration
	Timeouts              TimeoutsConfig
)

func SetConfig() {
	viper.SetConfigFile("config.yaml")
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

	Cutter = CutterConfig{
		URL:     viper.GetString(`cutter.url`),
		Timeout: viper.GetDuration(`cutter.timeout`),
	}

	Classifier = ClassifierConfig{
		URL: viper.GetString(`classifier.url`),
	}

	Similarity = SimilarityConfig{
		URL:     viper.GetString(`similarity.url`),
		Timeout: viper.GetDuration(`similarity.timeout`),
	}

	WellSimilarityPercent = viper.GetInt("well_similarity_percent")

	ExpirationCookieTime = viper.GetDuration("expiration_cookie_time")

	Timeouts = TimeoutsConfig{
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    5 * time.Second,
		ContextTimeout: time.Second * 2,
	}
}
