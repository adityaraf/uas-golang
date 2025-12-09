package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Achievement model untuk MongoDB
type Achievement struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AchievementID string             `bson:"achievement_id" json:"achievement_id"`
	StudentID     string             `bson:"student_id" json:"student_id"`
	Title         string             `bson:"title" json:"title"`
	Category      string             `bson:"category" json:"category"`
	Level         string             `bson:"level" json:"level"`
	Date          time.Time          `bson:"date" json:"date"`
	Description   string             `bson:"description" json:"description"`
	Documents     []Document         `bson:"documents" json:"documents"`
	Status        string             `bson:"status" json:"status"`
	IsDeleted     bool               `bson:"is_deleted" json:"is_deleted"`
	DeletedAt     *time.Time         `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// Document model untuk file upload
type Document struct {
	Filename   string    `bson:"filename" json:"filename"`
	Filepath   string    `bson:"filepath" json:"filepath"`
	Filesize   int64     `bson:"filesize" json:"filesize"`
	Mimetype   string    `bson:"mimetype" json:"mimetype"`
	UploadedAt time.Time `bson:"uploaded_at" json:"uploaded_at"`
}

// SubmitAchievementRequest untuk request body
type SubmitAchievementRequest struct {
	Title       string `json:"title" form:"title"`
	Category    string `json:"category" form:"category"`
	Level       string `json:"level" form:"level"`
	Date        string `json:"date" form:"date"` // Format: YYYY-MM-DD
	Description string `json:"description" form:"description"`
}

// AchievementResponse untuk response
type AchievementResponse struct {
	ID            string     `json:"id"`
	AchievementID string     `json:"achievement_id"`
	StudentID     string     `json:"student_id"`
	Title         string     `json:"title"`
	Category      string     `json:"category"`
	Level         string     `json:"level"`
	Date          time.Time  `json:"date"`
	Description   string     `json:"description"`
	Documents     []Document `json:"documents"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PaginationMeta untuk metadata pagination
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// AchievementListResponse untuk response list dengan pagination
type AchievementListResponse struct {
	Achievements []Achievement  `json:"achievements"`
	Pagination   PaginationMeta `json:"pagination"`
}