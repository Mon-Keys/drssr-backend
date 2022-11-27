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

type TGBotConfig struct {
	APIToken           string
	AdminChatID        int64
	EmailNotifications bool
}

type MailerConfig struct {
	Email    string
	Password string
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
	TgBotAPIToken         TGBotConfig
	Mailer                MailerConfig
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

	TgBotAPIToken = TGBotConfig{
		APIToken:           viper.GetString("tg_bot_stylist_accept.api_token"),
		AdminChatID:        viper.GetInt64("tg_bot_stylist_accept.admin_chat_id"),
		EmailNotifications: viper.GetBool("tg_bot_stylist_accept.email_notifications"),
	}

	Mailer = MailerConfig{
		Email:    viper.GetString("mailer.email"),
		Password: viper.GetString("mailer.password"),
	}

	WellSimilarityPercent = viper.GetInt("well_similarity_percent")

	ExpirationCookieTime = viper.GetDuration("expiration_cookie_time")

	Timeouts = TimeoutsConfig{
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    5 * time.Second,
		ContextTimeout: time.Second * 2,
	}
}
