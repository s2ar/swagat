package config

import (
	"github.com/spf13/viper"
)

type Server struct {
	Listen string
}

type BitmexService struct {
	APIKey    string
	APISecret string
	Verb      string
	Scheme    string
	Host      string
	Endpoint  string
}

type Configuration struct {
	BitmexService BitmexService
	Server        Server
}

func InitConfig(configFile string) (*Configuration, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.SetConfigFile(configFile)
	v.AutomaticEnv()
	v.AddConfigPath(".")

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	config := &Configuration{}

	err = v.Unmarshal(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
