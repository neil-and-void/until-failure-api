package config

import "time"

const (
	ACCESS_TTL  time.Duration = 720 // hours
	REFRESH_TTL time.Duration = 24  // hours

	// these are not the actual secrets, but are the keys to get the secrets
	// from the .env file
	ACCESS_SECRET  = "ACCESS_SECRET"
	REFRESH_SECRET = "REFRESH_SECRET"
	EMAIL          = "EMAIL"
	APP_PASSWORD   = "APP_PASSWORD"
)
