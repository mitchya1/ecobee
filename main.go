package main

import (
	"github.com/mitchya1/ecobee/pkg/ecobee"
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
	c := ecobee.GetOAuth(viper.GetString("api_token"), viper.GetString("auth_code"), viper.GetString("token_file"))
	//fmt.Printf("%+v", c)

	ecobee.GetThermostats(c.AccessToken)

}
