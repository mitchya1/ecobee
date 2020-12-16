package ecobee

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	thermostatURL = "https://api.ecobee.com/1/thermostat"
)

type thermostatReqProperties struct {
	Selection ecobeeSelection `json:"selection"`
}

// TODO add the rest from here https://www.ecobee.com/home/developer/api/documentation/v1/objects/Selection.shtml
type ecobeeSelection struct {
	SelectionType  string `json:"selectionType"`
	SelectionMatch string `json:"selectionMatch"`
	IncludeAlerts  bool   `json:"includeAlerts"`
	IncludeRuntime string `json:"includeRuntime"`
}

type thermostatsResponse struct {
	ThermostatList []thermostatList `json:"thermostatList"`
	Page           thermostatPages  `json:"page"`
	Status         thermostatStatus `json:"status"`
}

type thermostatStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type thermostatPages struct {
	Page       int `json:"page"`
	TotalPages int `json:"totalPages"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
}
type thermostatList struct {
	ID       string `json:"identifier"`
	Name     string `json:"name"`
	Revision string `json:"thermostatRev"`
}

// GetThermostats accepts an access_token (t)
func GetThermostats(t string) thermostatsResponse {

	// TODO don't panic here, just refresh the token
	/*if CheckTokenExpiration() {
		panic("Access token is expired")
	}*/
	client := &http.Client{}

	req, err := http.NewRequest("GET", thermostatURL, nil)

	if err != nil {
		panic(err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t))
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("cache-control", "no-cache")

	p := &thermostatReqProperties{
		Selection: ecobeeSelection{
			SelectionType:  "registered",
			SelectionMatch: "",
		},
	}

	r, err := json.Marshal(p)

	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("json", string(r))
	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())

	resp, _ := client.Do(req)

	rb, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(rb))

	tr := &thermostatsResponse{}

	// TODO check if we need to refresh

	err = json.Unmarshal(rb, tr)

	if err != nil {
		panic(err)
	}

	return *tr
}

func GetCurrentTemperature(id string) {

}
