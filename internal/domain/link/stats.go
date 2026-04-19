package link

import "time"

type Stats struct {
	ShortCode string
	Url       string
	Visits    int64
	CreatedAt time.Time
}
