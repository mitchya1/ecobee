package main

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
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

	var oauth ecobee.Tokens
	var thermostats ecobee.ThermostatsResponse

	c := influxdb2.NewClient(viper.GetString("influxdb_uri"), viper.GetString("influxdb_token"))
	defer c.Close()

	ic := ecobee.InfluxContainer{
		Client: c,
		Bucket: viper.GetString("influxdb_bucket"),
		Org:    viper.GetString("influxdb_org"),
	}

	// API key, authorization code
	oauth = ecobee.GetOAuth(viper.GetString("api_token"), viper.GetString("auth_code"), viper.GetString("token_file"))

	thermostats, err = ecobee.GetThermostats(oauth.AccessToken)

	if err != nil {
		if err == ecobee.ErrTokenExpired {
			oauth = ecobee.RefreshToken(viper.GetString("api_token"), viper.GetString("token_file"))
			thermostats, err = ecobee.GetThermostats(oauth.AccessToken)
		}
	}

	if err != nil {
		panic(err)
	}

	actual := thermostats.ThermostatList[0].Runtime.ActualTemperature
	dh := thermostats.ThermostatList[0].Runtime.DesiredHeat
	dc := thermostats.ThermostatList[0].Runtime.DesiredCool

	go ic.StoreTemperature(dh, dc, actual)
}
