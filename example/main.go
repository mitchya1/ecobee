package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	owm "github.com/briandowns/openweathermap"
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

	// Create InfluxContainer struct to hold influxdb config info
	// Also stores the client so we can write tests against it
	// This is optional
	ic := ecobee.InfluxContainer{
		Client: c,
		Bucket: viper.GetString("influxdb_bucket"),
		Org:    viper.GetString("influxdb_org"),
	}

	// Create HTTP client that will be used to make calls to the Ecobee API
	client := &http.Client{}

	// Populate the EBClient struct
	// GetOAuthTokens and RefreshTokens receive this struct
	ecobeeClient := ecobee.EBClient{
		Client:    client,
		APIKey:    viper.GetString("api_token"),
		AuthCode:  viper.GetString("auth_code"),
		TokenFile: viper.GetString("token_file"),
	}

	// GetOAuthTokens returns access and refresh tokens and also writes them to the token file specified in EBClient
	// This function will attempts to refresh the tokens if the access token is expired
	// If that fails, an error is returned and you should exit and check the logs
	oauth, err = ecobeeClient.GetOAuthTokens()

	if err != nil {
		os.Exit(1)
	}

	// This returns a list of thermostats along with their runtime information
	// Runtime information contains temperature information
	// Ecobee says not to use this for polling, but I haven't looked at their polling documentation enough to write code against it
	// Ecobee also says thermostats usually send data back to ecobee every 15 minutes, so don't call this more often than that
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

	w, err := owm.NewCurrent("F", "en", viper.GetString("owm_api_key")) // fahrenheit (imperial) with English output
	if err != nil {
		log.Fatalln(err)
	}

	if err = w.CurrentByZip(viper.GetInt("owm_zip_code"), "US"); err != nil {
		fmt.Println("Error getting weather")
	}

	// Storing data in influx is optional
	for _, therm := range thermostats.ThermostatList {
		ic.StoreTemperature(
			therm.Runtime.DesiredHeat,
			therm.Runtime.DesiredCool,
			therm.Runtime.ActualTemperature,
			therm.Name)

		if w.Cod == 200 {
			ic.StoreCurrentOutsideTemperature(w.Main.Temp, therm.Name)
		}
	}

}
