package service

import (
	"context"
	models "crud-app/app/model"
	"crud-app/app/repository"
	"crud-app/app/utils"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementService struct {
	achievementRepo *repository.AchievementRepository
	referenceRepo   *repository.AchievementReferenceRepository
	uploadConfig    utils.FileUploadConfig
}

func NewAchievementService(mongoDB *mongo.Database, postgresDB *sql.DB) *AchievementService {
	return &AchievementService{
		achievementRepo: repository.NewAchievementRepository(mongoDB),
		referenceRepo:   repository.NewAchievementReferenceRepository(postgresDB),
		uploadConfig:    utils.DefaultUploadConfig,
	}
}

// SubmitAchievement - FR-003: Submit Prestasi
func (s *AchievementService) SubmitAchievement(c *fiber.Ctx) error {
	// Step 1: Get user_id dari context (dari AuthRequired middleware)
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized: User ID tidak ditemukan",
		})
	}

	// Step 2: Parse form data
	var req models.SubmitAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validasi input
	if req.Title == "" || req.Category == "" || req.Level == "" || req.Date == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Title, category, level, dan date harus diisi",
		})
	}

	// Parse date
	achievementDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Format date tidak valid. Gunakan format YYYY-MM-DD",
		})
	}

	// Step 3: Handle file upload (dokumen pendukung)
	form, err := c.MultipartForm()
	var documents []models.Document

	if err == nil && form != nil {
		files := form.File["documents"]
		for _, file := range files {
			// Save file
			filepath, err := utils.SaveUploadedFile(file, s.uploadConfig)
			if err != nil {
				return c.Status(400).JSON(fiber.Map{
					"status":  "error",
					"message": fmt.Sprintf("Gagal upload file: %v", err),
				})
			}

			// Add to documents
			documents = append(documents, models.Document{
				Filename:   file.Filename,
				Filepath:   filepath,
				Filesize:   file.Size,
				Mimetype:   file.Header.Get("Content-Type"),
				UploadedAt: time.Now(),
			})
		}
	}

	// Generate achievement ID
	achievementID := uuid.New().String()

	// Step 4: Simpan ke MongoDB (full document)
	achievement := &models.Achievement{
		AchievementID: achievementID,
		StudentID:     userID,
		Title:         req.Title,
		Category:      req.Category,
		Level:         req.Level,
		Date:          achievementDate,
		Description:   req.Description,
		Documents:     documents,
		Status:        "draft", // Status awal: draft
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	ctx := context.Background()
	if err := s.achievementRepo.Create(ctx, achievement); err != nil {
		// Rollback: hapus uploaded files
		for _, doc := range documents {
			utils.DeleteFile(doc.Filepath)
		}
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menyimpan achievement ke MongoDB",
		})
	}

	// Step 5: Simpan reference ke PostgreSQL
	reference := &models.AchievementReferences{
		ID:                 uuid.New(),
		StudentID:          uuid.MustParse(userID),
		MongoAchievementID: achievementID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.referenceRepo.Create(reference); err != nil {
		// Rollback: hapus dari MongoDB dan files
		s.achievementRepo.Delete(ctx, achievementID)
		for _, doc := range documents {
			utils.DeleteFile(doc.Filepath)
		}
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menyimpan reference ke PostgreSQL",
		})
	}

	// Step 6: Return achievement data
	response := models.AchievementResponse{
		ID:            achievement.ID.Hex(),
		AchievementID: achievement.AchievementID,
		StudentID:     achievement.StudentID,
		Title:         achievement.Title,
		Category:      achievement.Category,
		Level:         achievement.Level,
		Date:          achievement.Date,
		Description:   achievement.Description,
		Documents:     achievement.Documents,
		Status:        achievement.Status,
		CreatedAt:     achievement.CreatedAt,
		UpdatedAt:     achievement.UpdatedAt,
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Prestasi berhasil disubmit",
		"data":    response,
	})
}

// GetMyAchievements mendapatkan semua prestasi milik user yang login
func (s *AchievementService) GetMyAchievements(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	ctx := context.Background()
	achievements, err := s.achievementRepo.FindByStudentID(ctx, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data achievements",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data achievements berhasil diambil",
		"data":    achievements,
	})
}

// GetAchievementByID mendapatkan detail achievement
func (s *AchievementService) GetAchievementByID(c *fiber.Ctx) error {
	achievementID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)

	ctx := context.Background()
	achievement, err := s.achievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement tidak ditemukan",
		})
	}

	// Check ownership (hanya bisa lihat achievement sendiri, kecuali admin)
	roleID, _ := c.Locals("role_id").(string)
	if achievement.StudentID != userID && roleID != "1" {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Anda tidak memiliki akses ke achievement ini",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data achievement berhasil diambil",
		"data":    achievement,
	})
}

// UpdateAchievement mengupdate achievement (hanya jika status masih draft)
func (s *AchievementService) UpdateAchievement(c *fiber.Ctx) error {
	achievementID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)

	ctx := context.Background()

	// Get existing achievement
	existing, err := s.achievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement tidak ditemukan",
		})
	}

	// Check ownership
	if existing.StudentID != userID {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Anda tidak memiliki akses ke achievement ini",
		})
	}

	// Check status (hanya bisa update jika masih draft)
	if existing.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement yang sudah disubmit tidak bisa diupdate",
		})
	}

	// Parse request
	var req models.SubmitAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Update fields
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Level != "" {
		existing.Level = req.Level
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Date != "" {
		date, err := time.Parse("2006-01-02", req.Date)
		if err == nil {
			existing.Date = date
		}
	}

	// Update di MongoDB
	if err := s.achievementRepo.Update(ctx, achievementID, existing); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate achievement",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement berhasil diupdate",
		"data":    existing,
	})
}

// DeleteAchievement menghapus achievement (hanya jika status masih draft)
func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
	achievementID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)

	ctx := context.Background()

	// Get existing achievement
	existing, err := s.achievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement tidak ditemukan",
		})
	}

	// Check ownership
	if existing.StudentID != userID {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Anda tidak memiliki akses ke achievement ini",
		})
	}

	// Check status
	if existing.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement yang sudah disubmit tidak bisa dihapus",
		})
	}

	// Delete files
	for _, doc := range existing.Documents {
		utils.DeleteFile(doc.Filepath)
	}

	// Delete from MongoDB
	if err := s.achievementRepo.Delete(ctx, achievementID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menghapus achievement dari MongoDB",
		})
	}

	// Delete from PostgreSQL
	if err := s.referenceRepo.Delete(achievementID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menghapus reference dari PostgreSQL",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement berhasil dihapus",
	})
}

// SubmitForVerification - FR-004: Submit untuk Verifikasi
func (s *AchievementService) SubmitForVerification(c *fiber.Ctx) error {
	achievementID := c.Params("id")
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	ctx := context.Background()

	// Step 1: Get existing achievement
	achievement, err := s.achievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement tidak ditemukan",
		})
	}

	// Check ownership
	if achievement.StudentID != userID {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Anda tidak memiliki akses ke achievement ini",
		})
	}

	// Precondition: Status harus 'draft'
	if achievement.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Hanya prestasi dengan status 'draft' yang bisa disubmit",
		})
	}

	// Step 2: Update status menjadi 'submitted' di MongoDB
	if err := s.achievementRepo.UpdateStatus(ctx, achievementID, "submitted"); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate status di MongoDB",
		})
	}

	// Step 3: Update status dan submitted_at di PostgreSQL
	if err := s.referenceRepo.UpdateSubmittedStatus(achievementID); err != nil {
		// Rollback MongoDB
		s.achievementRepo.UpdateStatus(ctx, achievementID, "draft")
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate status di PostgreSQL",
		})
	}

	// Step 4: Return updated status
	achievement.Status = "submitted"
	achievement.UpdatedAt = time.Now()

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Prestasi berhasil disubmit untuk verifikasi",
		"data": fiber.Map{
			"achievement_id": achievement.AchievementID,
			"status":         achievement.Status,
			"updated_at":     achievement.UpdatedAt,
		},
	})
}