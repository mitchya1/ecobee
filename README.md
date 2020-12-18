# ecobee

![Tests](https://github.com/mitchya1/ecobee/workflows/Tests/badge.svg)

A module to work with the Ecobee API.

This is a work in progress. Right now, it can handle:

- Retrieving and storing access and refresh tokens

- Refreshing tokens

- Retrieving a list of thermostats and their runtime information (which includes temperature)

- Optionally writing information to InfluxDB

## Using

You can `go get` this module:

`go get github.com/mitchya1/ecobee/pkg/ecobee`

Follow the steps in the **Ecobee Setup** section for steps on how to set up Ecobee. This involves some manual work.

**Important Note:** After completing the steps in **Ecobee Setup**, you have 900 seconds to retrieve your access and refresh tokens.

At this point, you should have information to retrieve your tokens.

The initial code to retrieve tokens could look like this:

```go
package main

import (
  "net/http"
  "os"
  "github.com/mitchya1/ecobee/pkg/ecobee"
)
func main() {

  client := &http.Client{}

  ecobeeClient := ecobee.EBClient{
		Client:    client,
		APIKey:    "YOUR API KEY",
		AuthCode:  "YOUR AUTH CODE",
		TokenFile: "PATH TO TOKEN FILE",
  }
  
  // This will automatically store the tokens wherever you specify in
  // ecobee.EBClient.TokenFile
  oauth, err = ecobeeClient.GetOAuthTokens()

	if err != nil {
		os.Exit(1)
	}
}
```

### Configuration File

If you plan on borrowing from `/example/main.go`, here are some notes about the configuration file:


The configuration file **must** be named `ecobee.yml` and **must** exist in either:

- The directory that you call the binary from (If the binary exists in `/opt/ecobee/ecobee` and you run it from `/var/tmp/`, the config file should be `/var/tmp/ecobee.yml`)

- `$HOME/.ecobee/`

- `/etc/ecobee/`

A yaml file with these keys:

```yml
api_token: "API token from the application - found under 'Developer'"
auth_code: "Authorization code from Step 1: https://www.ecobee.com/home/developer/api/examples/ex1.shtml"
token_file: "Path to store access and refresh tokens"
influxdb_token: "InfluxDB API token"
influxdb_bucket: "bucket name to store data in"
influxdb_org: "org name"
influxdb_uri: "URI to influxdb, including protocol and port"
owm_zip_code: 11111
owm_api_key: "Your ecobee API key"
```


## Ecobee Setup 

Create an ecobee developer account [here](https://www.ecobee.com/developers/)
  - 2FA must be disabled on your ecobee account

Log into ecobee

Go to the developer tab

Create an app

Authorize the app under "My Apps" in the Ecobee web app

Retrieve the API key from the app you created

Go through the first step [here](https://www.ecobee.com/home/developer/api/examples/ex1.shtml) to retrieve your authorization code
  - This code can only be used once, otherwise you'll have to redo these steps and re-add the app under "My Apps" in the Ecobee web app

## TODO

Rework refresh token flow to make it automatic
  - This should happen in `ecobee.GetOAuth`

Make thermostat retrieval more [variabalized](https://www.ecobee.com/home/developer/api/documentation/v1/objects/Selection.shtml)