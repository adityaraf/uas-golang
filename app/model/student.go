package models

import "time"

type Student struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	StudentID     string    `json:"student_id"`
	ProgramStudy  string    `json:"program_study"`
	AcademicYear  string    `json:"academic_year"`
	AdvisorID     string    `json:"advisor_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type StudentDetail struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	StudentID     string `json:"student_id"`
	FullName      string `json:"full_name"`
	ProgramStudy  string `json:"program_study"`
	AcademicYear  string `json:"academic_year"`
	AdvisorID     string `json:"advisor_id"`
	AdvisorName   string `json:"advisor_name"`
}