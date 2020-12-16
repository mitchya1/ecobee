package ecobee

import "errors"

var (
	// ErrTokenExpired is returned when the access token must be refreshed
	ErrTokenExpired = errors.New("Access Token expired")
)
