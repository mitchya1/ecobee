package ecobee

import "errors"

var (
	// ErrTokenExpired is returned when the access token must be refreshed
	ErrTokenExpired = errors.New("Access Token expired")
	// ErrRateLimited is returned when Ecobee tells us to slow down
	ErrRateLimited = errors.New("Rate limited by Ecobee")
	// ErrUnaccountedFor is returned when an error is returned that we haven't written code to handle
	ErrUnaccountedFor = errors.New("Unaccounted for error")
)
