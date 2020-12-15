package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	authURL  = "https://api.ecobee.com/authorize"
	tokenURL = "https://api.ecobee.com/token"
)

type ecobeeAuthorizationCodeResponse struct {
	EcobeePIN           string `json:"ecobeePin"`
	Code                string `json:"code"`
	Interval            int    `json:"interval"`
	ExpirationInSeconds int    `json:"expires_in"`
	Scope               string `json:"scope"`
}

type EcobeeOAuthResponse struct {
	AccessToken         string `json:"access_token"`
	TokenType           string `json:"token_type"`
	ExpirationInSeconds int64  `json:"expires_in"`
	Scope               string `json:"scope"`
	RefreshToken        string `json:"refresh_token"`
}

var (
	TokenExpirationEpoch int64
)

// getAuthorizationCode returns an authorization code, given an API key (k) and an ecobee app PIN (p)
// Ecobee docs: https://www.ecobee.com/home/developer/api/examples/ex1.shtml
func getAuthorizationCode(k string) (string, error) {

	client := &http.Client{}

	req, err := http.NewRequest("GET", authURL, nil)

	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("response_type", "ecobeePin")
	q.Add("client_id", k)
	q.Add("scope", "smartWrite") // Todo change to read, we don't need to write
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	rb, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(rb))

	e := &ecobeeAuthorizationCodeResponse{}

	err = json.Unmarshal(rb, e)

	if err != nil {
		panic(err)
	}

	return e.Code, nil

}

// GetOAuth returns a struct containing the OAuth response from Ecobee
func GetOAuth(k, c string) EcobeeOAuthResponse {

	/*
		c, err := getAuthorizationCode(k)

		if err != nil {
			fmt.Println("Error")
		}
	*/

	client := &http.Client{}

	req, err := http.NewRequest("GET", tokenURL, nil)

	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("grant_type", "ecobeePin")
	q.Add("code", c)
	q.Add("client_id", k)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	now := time.Now().Unix()

	if err != nil {
		panic(err)
	}

	rb, _ := ioutil.ReadAll(resp.Body)
	e := &EcobeeOAuthResponse{}

	err = json.Unmarshal(rb, e)

	if err != nil {
		panic(err)
	}

	TokenExpirationEpoch = now + e.ExpirationInSeconds

	return *e
}

func refreshToken(t string) {

}
