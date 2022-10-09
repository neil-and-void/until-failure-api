package config

import "time"

const (
	ACCESS_TTL     time.Duration = 720 // hours
	REFRESH_TTL    time.Duration = 24  // hours
	ACCESS_SECRET                = "ACCESS_SECRET"
	REFRESH_SECRET               = "REFRESH_SECRET"
)
