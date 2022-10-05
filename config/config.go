package config

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

type AppConfig struct {
	Env                    string `mapstructure:"ENV_NAME"`
	LogLevel               string `mapstructure:"LOG_LEVEL"`
	AccessKey              string `mapstructure:"ACCESS_KEY"`
	SenderId               string `mapstructure:"SENDER_COMPID"`
	Passphrase             string `mapstructure:"PASSPHRASE"`
	SigningKey             string `mapstructure:"SIGNING_KEY"`
	SessionKey             string `mapstructure:"SESSION_KEY"`
	PrimeApiUrl            string `mapstructure:"PRIME_API_URL"`
	PortfolioId            string `mapstructure:"PORTFOLIO_ID"`
	PrimeCredentials       string `mapstructure:"PRIME_CREDENTIALS"`
	PriceKinesisStreamName string `mapstructure:"PRICE_KDS_STREAM_NAME"`
	OrderKinesisStreamName string `mapstructure:"ORDER_KDS_STREAM_NAME"`
	AwsRegion              string `mapstructure:"AWS_REGION"`
}

func (a AppConfig) IsLocalEnv() bool {
	return a.Env == "local"
}

func Setup(app *AppConfig) error {

	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)

	// Set defaults
	viper.SetDefault("LOG_LEVEL", "warning")
	viper.SetDefault("PORT", "8443")
	viper.SetDefault("ENV_NAME", "local")

	err := viper.ReadInConfig()
	if err != nil {
		log.Infof("Missing env file %v", err)
	}

	err = viper.Unmarshal(&app)
	if err != nil {
		log.Infof("Cannot parse env file %v", err)
	}

	if app.IsLocalEnv() {
		return nil
	}

	// TODO: Why aren't these being picked up by viper? Do you have
	// to set a default?

	app.PrimeApiUrl = os.Getenv("PRIME_API_URL")
	app.AwsRegion = os.Getenv("AWS_REGION")
	app.PriceKinesisStreamName = os.Getenv("PRICE_KDS_STREAM_NAME")
	app.OrderKinesisStreamName = os.Getenv("ORDER_KDS_STREAM_NAME")
	app.PrimeCredentials = os.Getenv("PRIME_CREDENTIALS")

	// Parse the prime credentials
	var creds map[string]interface{}
	err = json.Unmarshal([]byte(app.PrimeCredentials), &creds)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal prime credentials: %v", err)
	}

	app.AccessKey = creds["accessKey"].(string)
	app.Passphrase = creds["passphrase"].(string)
	app.SigningKey = creds["signingKey"].(string)
	app.PortfolioId = creds["portfolioId"].(string)
	app.SenderId = creds["svcAccountId"].(string)

	return nil
}
