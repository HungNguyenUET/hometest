package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	FilePath           string `mapstructure:"FILE_PATH"`
	SummaryFilePath    string `mapstructure:"SUMMARY_FILE_PATH"`
	OpenAiLimitRequest int    `mapstructure:"OPEN_AI_LIMIT_REQUEST"`
	OpenAIKey          string `mapstructure:"OPEN_AI_KEY"`
	OpenAIURL          string `mapstructure:"OPEN_AI_URL"`
	OpenAIModel        string `mapstructure:"OPEN_AI_MODEL"`
}

var cfg *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using environment variables")
		} else {
			log.Fatalf("Error reading config file: %s", err)
		}
	}

	setDefaults()

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Unable to decode config into struct: %v", err)
	}
}

func setDefaults() {
	viper.SetDefault("FILE_PATH", "TheArtOfThinkingClearly.txt")
	viper.SetDefault("SUMMARY_FILE_PATH", "Summary.txt")
	viper.SetDefault("OPEN_AI_LIMIT_REQUEST", 5)
	viper.SetDefault("SUMMARY_FILE_PATH", "Summary.txt")
	viper.SetDefault("SUMMARY_FILE_PATH", "Summary.txt")
}

func GetConfig() *Config {
	return cfg
}
