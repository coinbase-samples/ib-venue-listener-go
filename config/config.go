package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Env         string `mapstructure:"ENV_NAME"`
	LogLevel    string `mapstructure:"LOG_LEVEL"`
	AccessKey   string `mapstructure:"ACCESS_KEY"`
	SenderId    string `mapstructure:"SENDER_COMPID"`
	Passphrase  string `mapstructure:"PASSPHRASE"`
	SigningKey  string `mapstructure:"SIGNING_KEY"`
	SessionKey  string `mapstructure:"SESSION_KEY"`
	PrimeApiUrl string `mapstructure:"PRIME_API_URL"`
	PortfolioId string `mapstructure:"PORTFOLIO_ID"`
}

func Setup(app *AppConfig) {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	// set defaults
	viper.SetDefault("LOG_LEVEL", "warning")
	viper.SetDefault("PORT", "8443")
	viper.SetDefault("ENV_NAME", "local")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Missing env file %v\n", err)
	}

	err = viper.Unmarshal(&app)
	if err != nil {
		fmt.Printf("Cannot parse env file %v\n", err)
	}
}
