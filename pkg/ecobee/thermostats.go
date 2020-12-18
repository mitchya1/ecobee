package ecobee

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/common/log"
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
	IncludeRuntime bool   `json:"includeRuntime"`
}

// ThermostatsResponse is the response from retrieving the ecobee thermostats in your account
type ThermostatsResponse struct {
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
	ID       string            `json:"identifier"`
	Name     string            `json:"name"`
	Revision string            `json:"thermostatRev"`
	Runtime  thermostatRuntime `json:"runtime"`
}

type thermostatRuntime struct {
	Connected bool `json:"connected"`
	// TODO get lastStatusModified, which is a time.Time value of some sort
	ActualTemperature int `json:"actualTemperature"`
	ActualHumidity    int `json:"actualHumidity"`
	DesiredHeat       int `json:"desiredHeat"`
	DesiredCool       int `json:"desiredCool"`
}

// GetThermostats accepts an access_token (t)
// If you receive an ErrTokenExpired, you must call ecobee.RefreshTokens() then retry your request
func GetThermostats(t string) (ThermostatsResponse, error) {

	tr := &ThermostatsResponse{}

	log.Info("Retrieving thermostats and their readings")

	client := &http.Client{}

	req, err := http.NewRequest("GET", thermostatURL, nil)

	if err != nil {
		return *tr, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t))
	req.Header.Add("Content-Type", "application/json;charset=UTF-8")
	req.Header.Add("cache-control", "no-cache")

	p := &thermostatReqProperties{
		Selection: ecobeeSelection{
			SelectionType:  "registered",
			SelectionMatch: "",
			IncludeRuntime: true,
		},
	}

	r, err := json.Marshal(p)

	if err != nil {
		log.Error("Error marshalling thermostatReqProperties")
		return *tr, err
	}

	q := req.URL.Query()
	q.Add("json", string(r))
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	if err != nil {
		log.Error("Error making HTTP request to retrieve thermostats. Error: ", err.Error())
		resp.Body.Close()
		return *tr, err
	}

	defer resp.Body.Close()

	rb, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error("Error reading response body from thermostat call. Error: ", err.Error())
		return *tr, err
	}

	err = json.Unmarshal(rb, tr)

	if err != nil {
		log.Error("Error unmarshaling response into ThermostatsResponse. Error: ", err.Error())
		return *tr, err
	}

	if tr.Status.Code == 14 {
		log.Warn("Access token is expired, you should refresh it")
		return *tr, ErrTokenExpired
	}

	return *tr, nil
}

// GetCurrentTemperature isn't used yet because we retrieve that from the runtime info in GetThermostats
func GetCurrentTemperature(id string, t string) {

}
