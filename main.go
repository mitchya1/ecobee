package main

import (
	"fmt"

	"github.com/mitchya1/ecobee/pkg/auth"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()

	if err != nil {
		panic(err)
	}

	// API key, authorization code
	c := auth.GetOAuth(viper.GetString("api_token"), viper.GetString("auth_code"))
	fmt.Println(c)

}
