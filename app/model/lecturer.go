package models

import "time"

type Lecturer struct {
	ID         string    `db:"id" json:"id"`
	UserID     string    `db:"user_id" json:"user_id"`
	LecturerID string    `db:"lecturer_id" json:"lecturer_id"`
	Department string    `db:"department" json:"department"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type LecturerDetail struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	LecturerID string `json:"lecturer_id"`
	FullName   string `json:"full_name"`
	Department string `json:"department"`
}
