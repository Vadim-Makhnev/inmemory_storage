package storage

import "time"

type Value struct {
	Data      string
	ExpiresAt *time.Time
}
