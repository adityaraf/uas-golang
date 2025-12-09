package models

import (
	"time"

	"github.com/google/uuid"
)

type AchievementReferences struct {
	ID                 uuid.UUID  `json:"id"`
	StudentID          uuid.UUID  `json:"student_id"`
	MongoAchievementID string     `json:"mongo_achievement_id"`
	Status             string     `json:"status"`
	SubmittedAt        *time.Time `json:"submitted_at"`
	VerifiedAt         *time.Time `json:"verified_at"`
	VerifiedBy         *uuid.UUID `json:"verified_by"`
	RejectionNote      *string    `json:"rejection_note"`
	DeletedAt          *time.Time `json:"deleted_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"` 	
}