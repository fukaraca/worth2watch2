package config

import (
	"github.com/spf13/viper"
	"log"
)

var GetEnv = setEnv()

func setEnv() *viper.Viper {
	sE := viper.New()
	sE.AddConfigPath("./config")
	sE.AddConfigPath("./../config")
	sE.SetConfigName("config")
	sE.SetConfigType("env")
	err := sE.ReadInConfig()
	if err != nil {
		log.Println("viper config.env loading err:", err)
	}
	return sE
}
