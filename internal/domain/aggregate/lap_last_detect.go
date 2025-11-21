package aggregate

import "time"

type LapLastDetect struct {
	LapId      string    `json:"lap_id" db:"lap_id"`
	LastGroup  int       `json:"last_group" db:"last_group"`
	LastDetect time.Time `json:"last_detect" db:"last_detect"`
}
