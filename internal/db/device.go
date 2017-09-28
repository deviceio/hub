package db

import "time"

type Device struct {
	ID           string
	Hostname     string
	Platform     string
	Architecture string
	LastActivity *time.Time
}
