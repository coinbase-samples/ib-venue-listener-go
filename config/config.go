package config

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type AppConfig struct {
	//Local Only
	LocalStackHostname string `mapstructure:"LOCALSTACK_HOSTNAME"`
	Env                string `mapstructure:"ENV_NAME"`
	LogLevel           string `mapstructure:"LOG_LEVEL"`
	LogToFile          string `mapstructure:"LOG_TO_FILE"`
	AccessKey          string `mapstructure:"ACCESS_KEY"`
	SenderId           string `mapstructure:"SENDER_COMPID"`
	Passphrase         string `mapstructure:"PASSPHRASE"`
	SigningKey         string `mapstructure:"SIGNING_KEY"`
	SessionKey         string `mapstructure:"SESSION_KEY"`
	PrimeApiUrl        string `mapstructure:"PRIME_API_URL"`
	PortfolioId        string `mapstructure:"PORTFOLIO_ID"`
	PrimeCredentials   string `mapstructure:"PRIME_CREDENTIALS"`
	AwsRegion          string `mapstructure:"AWS_REGION"`
	OrderFillQueueUrl  string `mapstructure:"ORDER_FILL_QUEUE_URL"`
	AssetTableName     string `mapstructure:"PRODUCT_PRICE_TABLE_NAME"`
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
	viper.SetDefault("LOCALSTACK_HOSTNAME", "localstack")
	viper.SetDefault("LOG_LEVEL", "warning")
	viper.SetDefault("LOG_TO_FILE", "false")
	viper.SetDefault("PORT", "8443")
	viper.SetDefault("ENV_NAME", "local")

	viper.SetDefault("AWS_REGION", "us-east-1")
	viper.SetDefault("ORDER_FILL_QUEUE_URL", "http://localhost:4566/000000000000/orderFillQueue.fifo")
	viper.SetDefault("PRIME_API_URL", "ws-feed.prime.coinbase.com")
	viper.SetDefault("PRODUCT_PRICE_TABLE_NAME", "Asset")

	err := viper.ReadInConfig()
	if err != nil {
		log.Infof("Missing env file %v", err)
	}

	err = viper.Unmarshal(&app)
	if err != nil {
		log.Infof("Cannot parse env file %v", err)
	}

	// If app is not local, pull prime credentials from secret manager
	if app.IsLocalEnv() {
		return nil
	}

	app.PrimeApiUrl = os.Getenv("PRIME_API_URL")
	app.PrimeCredentials = os.Getenv("PRIME_CREDENTIALS")

	// Parse the prime credentials
	var creds map[string]interface{}
	err = json.Unmarshal([]byte(app.PrimeCredentials), &creds)
	if err != nil {
		return fmt.Errorf("unable to unmarshal prime credentials: %v", err)
	}

	app.AccessKey = creds["accessKey"].(string)
	app.Passphrase = creds["passphrase"].(string)
	app.SigningKey = creds["signingKey"].(string)
	app.PortfolioId = creds["portfolioId"].(string)
	app.SenderId = creds["svcAccountId"].(string)

	return nil
}
