package entity

import "github.com/google/uuid"

type Detection struct {
	Id        int       `json:"id" db:"id"`
	GroupId   int       `json:"group_id" db:"group_id"`
	ImageUid  uuid.UUID `json:"image_uid" db:"image_uid"`
	Class     string    `json:"class" db:"class"`
	IsProblem bool      `json:"is_problem" db:"is_problem"`
}
