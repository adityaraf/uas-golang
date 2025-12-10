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
studentRepo     *repository.StudentRepository
uploadConfig    utils.FileUploadConfig
}

func NewAchievementService(mongoDB *mongo.Database, postgresDB *sql.DB) *AchievementService {
return &AchievementService{
achievementRepo: repository.NewAchievementRepository(mongoDB),
referenceRepo:   repository.NewAchievementReferenceRepository(postgresDB),
studentRepo:     repository.NewStudentRepository(postgresDB),
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
IsDeleted:     false,
DeletedAt:     nil,
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

// DeleteAchievement - FR-005: Hapus Prestasi (Soft Delete)
func (s *AchievementService) DeleteAchievement(c *fiber.Ctx) error {
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

// Precondition: Status harus 'draft'
if existing.Status != "draft" {
return c.Status(400).JSON(fiber.Map{
"status":  "error",
"message": "Hanya prestasi dengan status 'draft' yang bisa dihapus",
})
}

// Step 2: Soft delete di MongoDB
if err := s.achievementRepo.SoftDelete(ctx, achievementID); err != nil {
return c.Status(500).JSON(fiber.Map{
"status":  "error",
"message": "Gagal menghapus achievement di MongoDB",
})
}

// Step 3: Soft delete reference di PostgreSQL
if err := s.referenceRepo.SoftDelete(achievementID); err != nil {
// Rollback MongoDB (restore from soft delete)
// Note: Untuk production, buat fungsi Restore() jika diperlukan
return c.Status(500).JSON(fiber.Map{
"status":  "error",
"message": "Gagal menghapus reference di PostgreSQL",
})
}

// Step 4: Return success message
return c.Status(200).JSON(fiber.Map{
"status":  "success",
"message": "Prestasi berhasil dihapus",
})
}

// GetAdviseeAchievements - FR-006: View Prestasi Mahasiswa Bimbingan
func (s *AchievementService) GetAdviseeAchievements(c *fiber.Ctx) error {
userID, ok := c.Locals("user_id").(string)
if !ok || userID == "" {
return c.Status(401).JSON(fiber.Map{
"status":  "error",
"message": "Unauthorized",
})
}

// Parse pagination parameters
page := c.QueryInt("page", 1)
limit := c.QueryInt("limit", 10)
if page < 1 {
page = 1
}
if limit < 1 || limit > 100 {
limit = 10
}
offset := (page - 1) * limit

ctx := context.Background()

// Step 1: Get list student IDs dari tabel students where advisor_id
studentIDs, err := s.studentRepo.FindStudentIDsByAdvisorID(userID)
if err != nil {
return c.Status(500).JSON(fiber.Map{
"status":  "error",
"message": "Gagal mengambil data mahasiswa bimbingan",
})
}

// Check if advisor has students
if len(studentIDs) == 0 {
return c.Status(200).JSON(fiber.Map{
"status":  "success",
"message": "Tidak ada mahasiswa bimbingan",
"data": fiber.Map{
"achievements": []models.Achievement{},
"pagination": models.PaginationMeta{
Page:       page,
Limit:      limit,
TotalItems: 0,
TotalPages: 0,
},
},
})
}

// Step 2: Get achievements references dengan filter student_ids
references, total, err := s.referenceRepo.FindByStudentIDs(studentIDs, limit, offset)
if err != nil {
return c.Status(500).JSON(fiber.Map{
"status":  "error",
"message": "Gagal mengambil data achievement references",
})
}

// Check if no achievements found
if len(references) == 0 {
return c.Status(200).JSON(fiber.Map{
"status":  "success",
"message": "Tidak ada prestasi mahasiswa bimbingan",
"data": fiber.Map{
"achievements": []models.Achievement{},
"pagination": models.PaginationMeta{
Page:       page,
Limit:      limit,
TotalItems: total,
TotalPages: 0,
},
},
})
}

// Step 3: Extract achievement IDs untuk fetch dari MongoDB
achievementIDs := make([]string, len(references))
for i, ref := range references {
achievementIDs[i] = ref.MongoAchievementID
}

// Step 4: Fetch detail dari MongoDB
achievements, err := s.achievementRepo.FindByAchievementIDs(ctx, achievementIDs)
if err != nil {
return c.Status(500).JSON(fiber.Map{
"status":  "error",
"message": "Gagal mengambil detail achievements dari MongoDB",
})
}

// Calculate total pages
totalPages := int(total) / limit
if int(total)%limit > 0 {
totalPages++
}

// Step 5: Return list dengan pagination
return c.Status(200).JSON(fiber.Map{
"status":  "success",
"message": "Data prestasi mahasiswa bimbingan berhasil diambil",
"data": fiber.Map{
"achievements": achievements,
"pagination": models.PaginationMeta{
Page:       page,
Limit:      limit,
TotalItems: total,
TotalPages: totalPages,
},
},
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

// GetPendingVerification - FR-007: Get list achievement yang perlu diverifikasi
func (s *AchievementService) GetPendingVerification(c *fiber.Ctx) error {
	// Get pagination params
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	// Get pending achievements dari PostgreSQL
	references, total, err := s.referenceRepo.FindPendingVerification(limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data pending verification",
		})
	}

	// Get full data dari MongoDB
	ctx := context.Background()
	var achievements []models.Achievement

	for _, ref := range references {
		achievement, err := s.achievementRepo.FindByID(ctx, ref.MongoAchievementID)
		if err == nil && achievement != nil {
			achievements = append(achievements, *achievement)
		}
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data pending verification berhasil diambil",
		"data": fiber.Map{
			"achievements": achievements,
			"pagination": fiber.Map{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_page": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}

// ReviewAchievementDetail - FR-007: Dosen review detail prestasi
func (s *AchievementService) ReviewAchievementDetail(c *fiber.Ctx) error {
	achievementID := c.Params("id")

	ctx := context.Background()
	achievement, err := s.achievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement tidak ditemukan",
		})
	}

	// Get reference data untuk info tambahan
	reference, err := s.referenceRepo.FindByMongoID(achievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil reference data",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data achievement berhasil diambil",
		"data": fiber.Map{
			"achievement": achievement,
			"reference":   reference,
		},
	})
}

// ApproveAchievement - FR-007: Dosen approve prestasi
func (s *AchievementService) ApproveAchievement(c *fiber.Ctx) error {
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

	// Check status (hanya bisa approve jika status submitted)
	if existing.Status != "submitted" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Hanya achievement dengan status 'submitted' yang bisa diapprove",
		})
	}

	// Update status di MongoDB
	if err := s.achievementRepo.UpdateStatus(ctx, achievementID, "verified"); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate status di MongoDB",
		})
	}

	// Update verification di PostgreSQL
	if err := s.referenceRepo.UpdateVerification(achievementID, userID, "verified"); err != nil {
		// Rollback MongoDB
		s.achievementRepo.UpdateStatus(ctx, achievementID, "submitted")
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate verification di PostgreSQL",
		})
	}

	// Get updated data
	updated, _ := s.achievementRepo.FindByID(ctx, achievementID)
	reference, _ := s.referenceRepo.FindByMongoID(achievementID)

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement berhasil diverifikasi",
		"data": fiber.Map{
			"achievement": updated,
			"reference":   reference,
		},
	})
}

// RejectAchievement - FR-007: Dosen reject prestasi
func (s *AchievementService) RejectAchievement(c *fiber.Ctx) error {
	achievementID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)

	// Parse request body untuk rejection note
	var req struct {
		RejectionNote string `json:"rejection_note"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if req.RejectionNote == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Rejection note harus diisi",
		})
	}

	ctx := context.Background()

	// Get existing achievement
	existing, err := s.achievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement tidak ditemukan",
		})
	}

	// Check status (hanya bisa reject jika status submitted)
	if existing.Status != "submitted" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Hanya achievement dengan status 'submitted' yang bisa direject",
		})
	}

	// Update status di MongoDB
	if err := s.achievementRepo.UpdateStatus(ctx, achievementID, "rejected"); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate status di MongoDB",
		})
	}

	// Update rejection di PostgreSQL
	if err := s.referenceRepo.UpdateRejection(achievementID, userID, req.RejectionNote); err != nil {
		// Rollback MongoDB
		s.achievementRepo.UpdateStatus(ctx, achievementID, "submitted")
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate rejection di PostgreSQL",
		})
	}

	// Get updated data
	updated, _ := s.achievementRepo.FindByID(ctx, achievementID)
	reference, _ := s.referenceRepo.FindByMongoID(achievementID)

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Achievement berhasil direject",
		"data": fiber.Map{
			"achievement": updated,
			"reference":   reference,
		},
	})
}