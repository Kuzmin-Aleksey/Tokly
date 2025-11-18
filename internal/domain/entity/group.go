package entity

import "time"

type Group struct {
	Id       int       `json:"id" db:"id"`
	LapId    int       `json:"lap_id" db:"lap_id"`
	CreateAt time.Time `json:"create_at" db:"create_at"`
}
