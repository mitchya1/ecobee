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

// Tokens stores our access and refresh tokens
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// EBClient holds all the information we need to make requests to the Ecobee APIs
// It also allows us to write tests because we can provide the http.Client
type EBClient struct {
	Client    *http.Client
	APIKey    string
	TokenFile string
	AuthCode  string
}

// GetOAuthTokens returns a struct containing the OAuth response from Ecobee
func (client EBClient) GetOAuthTokens() (Tokens, error) {
	var t Tokens
	var err error

	if checkTokensFileExists(client.TokenFile) {
		t, err = readTokensFromFile(client.TokenFile)

		if err != nil {
			log.Fatal("Unable to read tokens, therefore unable to proceed. Error: ", err.Error())
			os.Exit(1)
		}
		return t, nil
	}

	req, err := http.NewRequest("POST", tokenURL, nil)

	if err != nil {
		log.Error("Error creating request to retrieve tokens")
		return t, err
	}

	q := req.URL.Query()
	q.Add("grant_type", "ecobeePin")
	q.Add("code", client.AuthCode)
	q.Add("client_id", client.APIKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Client.Do(req)

	if err != nil {
		log.Error("Error making HTTP request to retrieve tokens")
		resp.Body.Close()
		return t, err
	}

	defer resp.Body.Close()

	rb, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Error("Error reading OAuth response body")
		return t, err
	}

	e := &OAuthResponse{}

	err = json.Unmarshal(rb, e)

	if e.AccessToken == "" {
		er := &AuthErrorResponse{}
		err = json.Unmarshal(rb, er)
		if err != nil {
			log.Error("Error unmarshaling oauth response into AuthErrorResponse")
			log.Error("Response body is: ", string(rb))
			return t, err
		}
		// Check the AuthErrorResponse.Error
		if er.Error == "invalid_grant" {
			log.Error("Received invalid grant response. Attempting to refresh tokens")
			client.RefreshToken()
		} else if er.Error == "slow_down" {
			log.Error("Ecobee is rate limiting. Dying")
			return t, ErrRateLimited
		} else {
			log.Error("Received unaccounted for error from ecobee API ", er.Error)
			return t, ErrUnaccountedFor
		}
	}

	if err != nil {
		log.Error("Error unmarshaling oauth response into OAuthResponse")
		return t, err
	}

	log.Info("Access token expires in ", e.ExpirationInSeconds, " seconds")

	t.AccessToken = e.AccessToken
	t.RefreshToken = e.RefreshToken

	writeTokens(e.AccessToken, e.RefreshToken, client.TokenFile)

	return t, nil
}

// RefreshToken returns a new OAuthResponse
// It accepts an API key (k) and a file location (loc)
func (client EBClient) RefreshToken() Tokens {

	var tok Tokens

	log.Info("Refreshing tokens")

	existingTokens, err := readTokensFromFile(client.TokenFile)

	if err != nil {
		log.Fatal("Unable to read tokens, therefore unable to proceed", err.Error())
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", tokenURL, nil)

	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Add("grant_type", "refresh_token")
	q.Add("code", existingTokens.RefreshToken)
	q.Add("client_id", client.APIKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Client.Do(req)

	if err != nil {
		log.Error("Error making HTTP request to refresh tokens")
		resp.Body.Close()
		panic(err)
	}

	defer resp.Body.Close()

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

	writeTokens(e.AccessToken, e.RefreshToken, client.TokenFile)

	return tok
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
func checkTokensFileExists(loc string) bool {

	log.Info("Checking if tokens file exists")

	if _, err := os.Stat(loc); err == nil {
		log.Info("Tokens file exists")
		return true
	}

	log.Error("File '", loc, "' does not exist")
	return false
}

// Reads in existing access and refresh tokens and returns them to caller
func readTokensFromFile(loc string) (Tokens, error) {

	t := &Tokens{}

	data, err := ioutil.ReadFile(loc)

	if err != nil {
		log.Error("Error reading tokens from file at", loc)
		return *t, err
	}

	err = json.Unmarshal(data, t)

	if err != nil {
		log.Error("Error unmarshaling bytes from file into Tokens")
		return *t, err
	}

	return *t, nil
}

type ecobeeAuthorizationCodeResponse struct {
	EcobeePIN           string `json:"ecobeePin"`
	Code                string `json:"code"` // This is use once
	Interval            int    `json:"interval"`
	ExpirationInSeconds int    `json:"expires_in"`
	Scope               string `json:"scope"`
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
