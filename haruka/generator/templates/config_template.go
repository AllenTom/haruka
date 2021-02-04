package templates

var ConfigTemplate = `package application

import "github.com/spf13/viper"

var AppConfig Config

type Config struct {
	Addr        string $1json:"addr"$1

}

func ReadConfig() error {
	configer := viper.New()
	configer.AddConfigPath("./")
	configer.SetConfigType("yaml")
	configer.SetConfigName("config")
	err := configer.ReadInConfig()
	if err != nil {
		return err
	}
	configer.SetDefault("addr", ":7600")
	AppConfig = Config{
		Addr:        configer.GetString("addr"),

	}
	return nil
}
`

var ConfigFileTemplate = `addr: 7600
`
