package ecobee

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/prometheus/common/log"
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

// OAuthResponse is the initial OAuth response ecobee sends back, it contains access and refresh tokens
// along with some other info we don't use yet
type OAuthResponse struct {
	AccessToken         string `json:"access_token"`
	TokenType           string `json:"token_type"`
	ExpirationInSeconds int64  `json:"expires_in"`
	Scope               string `json:"scope"`
	RefreshToken        string `json:"refresh_token"`
}

// AuthErrorResponse is returned by ecobee when there's an issue authenticating
// Usually this means the token is expired or you need to reauthenticate the app
type AuthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// getAuthorizationCode returns an authorization code, given an API key (k) and an ecobee app PIN (p)
// Ecobee docs: https://www.ecobee.com/home/developer/api/examples/ex1.shtml
// There shouldn't be a reason to call this function but I'm keeping it in here for future reference
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
func GetOAuth(k, c, loc string) Tokens {
	var t Tokens

	if checkExistingTokens(loc) {
		t = readTokensFromFile(loc)
		return t
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", tokenURL, nil)

	if err != nil {
		log.Error("Error creating request to retrieve tokens")
		panic(err)
	}

	q := req.URL.Query()
	q.Add("grant_type", "ecobeePin")
	q.Add("code", c)
	q.Add("client_id", k)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	if err != nil {
		log.Error("Error making HTTP request to retrieve tokens")
		panic(err)
	}

	rb, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error("Error reading OAuth response body")
		panic(err)
	}

	e := &OAuthResponse{}

	err = json.Unmarshal(rb, e)

	if e.AccessToken == "" {
		er := &AuthErrorResponse{}
		err = json.Unmarshal(rb, er)
		if err != nil {
			log.Error("Error unmarshaling oauth response into AuthErrorResponse")
			panic(err)
		}

		if er.Error == "invalid_grant" {
			RefreshToken(t.RefreshToken, loc)
		} else if er.Error == "slow_down" {
			log.Error("Ecobee is rate limiting. Dying")
			os.Exit(1)
		}
	}

	if err != nil {
		log.Error("Error unmarshaling oauth response into OAuthResponse")
		panic(err)
	}

	t.AccessToken = e.AccessToken
	t.RefreshToken = e.RefreshToken

	writeTokens(e.AccessToken, e.RefreshToken, loc)

	return t
}

// RefreshToken returns a new OAuthResponse
// It accepts an API key (k) and a file location (loc)
func RefreshToken(k, loc string) Tokens {

	log.Info("Refreshing tokens")

	existingTokens := readTokensFromFile(loc)

	var tok Tokens

	client := &http.Client{}

	req, err := http.NewRequest("POST", tokenURL, nil)

	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("grant_type", "refresh_token")
	q.Add("code", existingTokens.RefreshToken)
	q.Add("client_id", k)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	if err != nil {
		log.Error("Error making HTTP request to refresh tokens")
		panic(err)
	}

	rb, _ := ioutil.ReadAll(resp.Body)
	e := &OAuthResponse{}

	err = json.Unmarshal(rb, e)

	if err != nil {
		log.Error("Error unmarshaling response into OAuthResponse")
		panic(err)
	}

	if e.AccessToken == "" {
		log.Debug(string(rb))
		panic("Unable to refresh Tokens")
	}

	tok.AccessToken = e.AccessToken
	tok.RefreshToken = e.RefreshToken

	writeTokens(e.AccessToken, e.RefreshToken, loc)

	return tok
}

// Tokens stores our access and refresh tokens
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func writeTokens(at, rt, loc string) {
	log.Info("Storing tokens in tokenfile at ", loc)
	t := &Tokens{
		AccessToken:  at,
		RefreshToken: rt,
	}

	tb, err := json.Marshal(t)

	if err != nil {
		log.Error("Error marshaling tokens into Tokens")
		panic(err)
	}

	err = ioutil.WriteFile(loc, tb, 0644)

	if err != nil {
		log.Error("Error writing tokens file to ", loc)
		panic(err)
	}
}

// Returns true if file exists
func checkExistingTokens(loc string) bool {

	log.Info("Checking if tokens file exists")

	if _, err := os.Stat(loc); err == nil {
		return true
	}
	return false
}

func readTokensFromFile(loc string) Tokens {
	data, err := ioutil.ReadFile(loc)

	if err != nil {
		log.Error("Error reading tokens from file at", loc)
		panic(err)
	}

	t := &Tokens{}

	err = json.Unmarshal(data, t)

	if err != nil {
		panic(err)
	}

	return *t
}
