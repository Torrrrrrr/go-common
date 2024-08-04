package util

import (
	"database/sql"
	"time"
)

func TimeToNullTime(t time.Time) *sql.NullTime {
	return &sql.NullTime{
		Time:  t,
		Valid: true,
	}
}
