package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func readConfig() map[string]interface{} {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	} else {
		fmt.Println(viper.AllSettings())
	}
	return viper.AllSettings()
}
