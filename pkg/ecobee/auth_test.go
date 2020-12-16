package ecobee

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/assert"
)

const (
	oauthResponse = `{
			"access_token": "access",
			"refresh_token": "refresh",
			"expires_in": 86400,
			"scope": "scope,scope_again",
			"token_type": "pin"
		}`
)

func TestCheckTokensFileExists(t *testing.T) {
	exists := checkTokensFileExists("/tmp/foobarwhizbang123")

	if exists {
		t.FailNow()
	} else {
		log.Info("Received expected response 'false'")
	}

	exists = checkTokensFileExists("/home")

	if !exists {
		t.Failed()
	} else {
		log.Info("Received expected response 'true'")
	}

}

func TestReadTokensFile(t *testing.T) {
	d := []byte("{\"access_token\": \"some_token\", \"refresh_token\": \"another_token\"}")

	err := ioutil.WriteFile("/tmp/token", d, 0444)

	if err != nil {
		log.Error("Unable to write token file", err.Error())
		t.Failed()
	}

	tok, err := readTokensFromFile("/tmp/token")

	os.Remove("/tmp/token")

	if err != nil {
		log.Error("Error received from readTokensFromFile", err.Error())
		t.Failed()
	}

	if tok.AccessToken != "some_token" {
		log.Error("Access token does not match expected value")
		t.Failed()
	} else {
		log.Info("Access token matches expected value")
	}

	if tok.RefreshToken != "another_token" {
		log.Error("Refresh token does not match expected value")
		t.Failed()
	} else {
		log.Info("Access token matches expected value")
	}
}

func TestOAuthLogin(t *testing.T) {

	if _, err := os.Stat("tokens.yml"); err == nil {
		os.Remove("tokens.yml")
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "ecobeePin", r.URL.Query().Get("grant_type"))
		assert.Equal(t, "ciCode", r.URL.Query().Get("code"))
		assert.Equal(t, "ciAPIKey", r.URL.Query().Get("client_id"))
		w.Write([]byte(oauthResponse))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	ebClient := EBClient{
		Client:    httpClient,
		APIKey:    "ciAPIKey",
		AuthCode:  "ciCode",
		TokenFile: "./ci-tokens.yml",
	}

	_, err := ebClient.GetOAuthTokens()

	if err != nil {
		t.Fail()
	}

}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewTLSServer(handler)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return client, s.Close
}
