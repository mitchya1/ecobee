package ecobee

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const (
	authURL  = "https://api.ecobee.com/authorize"
	tokenURL = "https://api.ecobee.com/token"
)

type ecobeeAuthorizationCodeResponse struct {
	EcobeePIN           string `json:"ecobeePin"`
	Code                string `json:"code"` // This is use once
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

type EcobeeAuthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
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
func GetOAuth(k, c, loc string) tokens {

	var t tokens

	if checkExistingTokens(loc) {
		t = readTokensFromFile(loc)
		return t
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", tokenURL, nil)

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

	if e.AccessToken == "" {
		er := &EcobeeAuthErrorResponse{}
		err = json.Unmarshal(rb, er)
		if err != nil {
			panic(err)
		}

		if er.Error == "invalid_grant" {
			RefreshToken(k, t.RefreshToken, loc)
		} else if er.Error == "slow_down" {
			fmt.Println("Ecobee is rate limiting. Dying")
			os.Exit(1)
		}
	}

	if err != nil {
		panic(err)
	}

	TokenExpirationEpoch = now + e.ExpirationInSeconds

	t.AccessToken = e.AccessToken
	t.RefreshToken = e.RefreshToken

	writeTokens(e.AccessToken, e.RefreshToken, loc)

	return t
}

// RefreshToken returns a new EcobeeOAuthResponse
// It accepts a refresh_token (t) and API key (k)
func RefreshToken(t, k, loc string) tokens {

	var tok tokens

	client := &http.Client{}

	req, err := http.NewRequest("POST", tokenURL, nil)

	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("grant_type", "refresh_token")
	q.Add("code", t)
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

	if e.AccessToken == "" {
		fmt.Println(string(rb))
		panic("Unable to refresh tokens")
	}

	TokenExpirationEpoch = now + e.ExpirationInSeconds

	tok.AccessToken = e.AccessToken
	tok.RefreshToken = e.RefreshToken

	writeTokens(e.AccessToken, e.RefreshToken, loc)

	return tok
}

// CheckTokenExpiration checks if the access_token from ecobee is expired
func CheckTokenExpiration() bool {
	if time.Now().Unix() >= TokenExpirationEpoch {
		return true
	}

	return false
}

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func writeTokens(at, rt, loc string) {
	t := &tokens{
		AccessToken:  at,
		RefreshToken: rt,
	}

	tb, err := json.Marshal(t)

	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(loc, tb, 0644)

	if err != nil {
		panic(err)
	}
}

// Returns true if file exists
func checkExistingTokens(loc string) bool {

	if _, err := os.Stat(loc); err == nil {
		return true
	}
	return false
}

func readTokensFromFile(loc string) tokens {
	data, err := ioutil.ReadFile(loc)

	if err != nil {
		panic(err)
	}

	t := &tokens{}

	err = json.Unmarshal(data, t)

	if err != nil {
		panic(err)
	}

	return *t
}
