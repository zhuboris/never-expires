package authservice

import "time"

const (
	AuthLifetime    = 60 * time.Minute
	RefreshLifetime = 180 * days
)

const (
	hoursInDay = 24
	days       = hoursInDay * time.Hour
)
