package main

import (
	"net/http"
	"os"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/mitchya1/ecobee/pkg/ecobee"
	"github.com/spf13/viper"
)

func main() {
	var err error
	var oauth ecobee.Tokens
	var thermostats ecobee.ThermostatsResponse

	viper.SetConfigName("ecobee")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/ecobee")
	viper.AddConfigPath("$HOME/.ecobee")

	err = viper.ReadInConfig()

	if err != nil {
		panic(err)
	}

	c := influxdb2.NewClient(viper.GetString("influxdb_uri"), viper.GetString("influxdb_token"))
	defer c.Close()

	ic := ecobee.InfluxContainer{
		Client: c,
		Bucket: viper.GetString("influxdb_bucket"),
		Org:    viper.GetString("influxdb_org"),
	}

	client := &http.Client{}

	ecobeeClient := ecobee.EBClient{
		Client:    client,
		APIKey:    viper.GetString("api_token"),
		AuthCode:  viper.GetString("auth_code"),
		TokenFile: viper.GetString("token_file"),
	}

	// API key, authorization code
	oauth, err = ecobeeClient.GetOAuthTokens()

	if err != nil {
		os.Exit(1)
	}

	thermostats, err = ecobee.GetThermostats(oauth.AccessToken)

	if err != nil {
		if err == ecobee.ErrTokenExpired {
			oauth = ecobeeClient.RefreshToken()
			thermostats, err = ecobee.GetThermostats(oauth.AccessToken)

			if err != nil {
				panic(err)
			}
		} else {
			os.Exit(1)
		}
	}

	actual := thermostats.ThermostatList[0].Runtime.ActualTemperature
	dh := thermostats.ThermostatList[0].Runtime.DesiredHeat
	dc := thermostats.ThermostatList[0].Runtime.DesiredCool

	ic.StoreTemperature(dh, dc, actual)
}
