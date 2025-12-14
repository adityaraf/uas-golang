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
	"go.mongodb.org/mongo-driver/bson"
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

// SubmitAchievement godoc
// @Summary Submit new achievement
// @Description Student submits a new achievement with supporting documents. Uses hybrid database storage (MongoDB + PostgreSQL).
// @Tags Achievements
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param title formData string true "Achievement title"
// @Param category formData string true "Achievement category (e.g., Academic Competition, Research, etc.)"
// @Param level formData string true "Achievement level (Local/Regional/National/International)"
// @Param date formData string true "Achievement date (YYYY-MM-DD format)"
// @Param description formData string false "Detailed achievement description"
// @Param documents formData file false "Supporting documents (certificates, photos, etc. - multiple files allowed)"
// @Success 201 {object} object{status=string,message=string,data=models.Achievement} "Achievement created successfully with draft status"
// @Failure 400 {object} map[string]interface{} "Invalid request, missing required fields, or file upload error"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires achievements.create)"
// @Failure 500 {object} map[string]interface{} "Internal server error - database or file system error"
// @Router /achievements [post]
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

// GetMyAchievements godoc
// @Summary Get my achievements
// @Description Get all achievements belonging to the authenticated user (all statuses included).
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{status=string,message=string,data=[]models.Achievement} "Achievements retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires achievements.read)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve achievements from database"
// @Router /achievements [get]
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

// GetAchievementByID godoc
// @Summary Get achievement by ID
// @Description Get detailed information about a specific achievement with access control (owner, admin, or lecturer only).
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} object{status=string,message=string,data=models.Achievement} "Achievement retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Access denied - not owner, admin, or lecturer"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve achievement data"
// @Router /achievements/{id} [get]
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

// UpdateAchievement godoc
// @Summary Update achievement
// @Description Update achievement information (only if status is draft). Validates ownership and status before updating.
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Param request body models.SubmitAchievementRequest true "Achievement update request"
// @Success 200 {object} object{status=string,message=string,data=models.Achievement} "Achievement updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or achievement cannot be updated (not draft status)"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Access denied - not owner or insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Update operation failed"
// @Router /achievements/{id} [put]
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

// DeleteAchievement godoc
// @Summary Delete achievement
// @Description Soft delete an achievement (only if status is draft). Validates ownership and status before deletion.
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} object{status=string,message=string} "Achievement deleted successfully"
// @Failure 400 {object} map[string]interface{} "Achievement cannot be deleted (not draft status)"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Access denied - not owner or insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Delete operation failed"
// @Router /achievements/{id} [delete]
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

// GetAdviseeAchievements godoc
// @Summary Get advisee achievements
// @Description Lecturer gets paginated list of achievements from their advisees (students under supervision).
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" default(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" default(10)
// @Success 200 {object} object{status=string,message=string,data=object{achievements=[]models.Achievement,pagination=models.PaginationMeta}} "Advisee achievements retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires lecturer access)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve achievements"
// @Router /achievements/advisees [get]
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

// SubmitForVerification godoc
// @Summary Submit achievement for verification
// @Description Submit a draft achievement for verification by lecturer. Changes status from 'draft' to 'submitted'.
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} object{status=string,message=string,data=object{achievement_id=string,status=string,updated_at=string}} "Achievement submitted successfully"
// @Failure 400 {object} map[string]interface{} "Achievement cannot be submitted (not draft status)"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Access denied - not owner or insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Submission process failed - database error"
// @Router /achievements/{id}/submit [post]
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

// GetPendingVerification godoc
// @Summary Get pending verification achievements
// @Description Get paginated list of achievements that are pending verification (submitted status).
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" default(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" default(10)
// @Success 200 {object} object{status=string,message=string,data=object{achievements=[]models.Achievement,pagination=object}} "Pending achievements retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires achievements.verify)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve pending achievements"
// @Router /achievements/pending [get]
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

// ReviewAchievementDetail godoc
// @Summary Review achievement detail
// @Description Lecturer reviews detailed information of an achievement for verification process.
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} object{status=string,message=string,data=object{achievement=models.Achievement,reference=object}} "Achievement details retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires achievements.verify)"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve achievement details"
// @Router /achievements/{id}/review [get]
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

// ApproveAchievement godoc
// @Summary Approve achievement
// @Description Lecturer approves a submitted achievement. Changes status from 'submitted' to 'verified'.
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} object{status=string,message=string,data=object{achievement=models.Achievement,reference=object}} "Achievement approved successfully"
// @Failure 400 {object} map[string]interface{} "Achievement cannot be approved (not submitted status)"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires achievements.verify)"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Verification process failed - database error"
// @Router /achievements/{id}/verify [post]
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

// RejectAchievement godoc
// @Summary Reject achievement
// @Description Lecturer rejects a submitted achievement with reason. Changes status from 'submitted' to 'rejected'.
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Param request body object{rejection_note=string} true "Rejection request with reason"
// @Success 200 {object} object{status=string,message=string,data=object{achievement=models.Achievement,reference=object}} "Achievement rejected successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request, missing rejection note, or achievement cannot be rejected"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires achievements.verify)"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Rejection process failed - database error"
// @Router /achievements/{id}/reject [post]
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

// GetAllAchievements godoc
// @Summary Get all achievements (Admin)
// @Description Admin gets paginated list of all achievements with filtering and sorting options.
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" default(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" default(10)
// @Param status query string false "Filter by status (draft, submitted, verified, rejected)"
// @Param student_id query string false "Filter by student ID"
// @Param sort_by query string false "Sort by field (default: created_at)" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} object{status=string,message=string,data=object{achievements=[]models.Achievement,pagination=object,filters=object}} "All achievements retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid filter parameters"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires admin access)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve achievements"
// @Router /achievements/all [get]
func (s *AchievementService) GetAllAchievements(c *fiber.Ctx) error {
	// Parse query parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	statusFilter := c.Query("status", "")
	studentIDFilter := c.Query("student_id", "")
	sortBy := c.Query("sort_by", "created_at")
	sortOrder := c.Query("sort_order", "desc")

	// Validation
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Validate status filter
	if statusFilter != "" {
		validStatuses := map[string]bool{
			"draft":     true,
			"submitted": true,
			"verified":  true,
			"rejected":  true,
		}
		if !validStatuses[statusFilter] {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid status filter. Valid values: draft, submitted, verified, rejected",
			})
		}
	}

	// Step 1: Get achievement references dari PostgreSQL dengan filter
	references, total, err := s.referenceRepo.FindAllWithFilters(
		limit,
		offset,
		statusFilter,
		studentIDFilter,
		sortBy,
		sortOrder,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data achievement references",
		})
	}

	// Check if no data
	if len(references) == 0 {
		return c.Status(200).JSON(fiber.Map{
			"status":  "success",
			"message": "Data achievements berhasil diambil",
			"data": fiber.Map{
				"achievements": []models.Achievement{},
				"pagination": fiber.Map{
					"page":        page,
					"limit":       limit,
					"total_items": total,
					"total_pages": 0,
				},
			},
		})
	}

	// Step 2: Extract achievement IDs untuk fetch dari MongoDB
	achievementIDs := make([]string, len(references))
	for i, ref := range references {
		achievementIDs[i] = ref.MongoAchievementID
	}

	// Step 3: Fetch details dari MongoDB
	ctx := context.Background()
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

	// Step 4: Return dengan pagination
	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data achievements berhasil diambil",
		"data": fiber.Map{
			"achievements": achievements,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total_items": total,
				"total_pages": totalPages,
			},
			"filters": fiber.Map{
				"status":     statusFilter,
				"student_id": studentIDFilter,
				"sort_by":    sortBy,
				"sort_order": sortOrder,
			},
		},
	})
}

// GetMyStatistics godoc
// @Summary Get my achievement statistics
// @Description Get comprehensive achievement statistics for the authenticated student including summary, category breakdown, level distribution, and period analysis.
// @Tags Statistics & Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{status=string,message=string,data=object} "Statistics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires achievements.read)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve statistics from database"
// @Router /reports/statistics [get]
func (s *AchievementService) GetMyStatistics(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	ctx := context.Background()

	// Get statistics dari MongoDB
	stats, err := s.achievementRepo.GetStatisticsByStudentIDs(ctx, []string{userID})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil statistik",
		})
	}

	// Build response
	response := buildStatisticsResponse(stats, false)

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Statistik berhasil diambil",
		"data":    response,
	})
}

// GetAdviseeStatistics godoc
// @Summary Get advisee statistics
// @Description Lecturer view of comprehensive achievement statistics for their advisees including top performers ranking.
// @Tags Statistics & Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{status=string,message=string,data=object} "Advisee statistics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires lecturer access)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve statistics from database"
// @Router /reports/statistics [get]
func (s *AchievementService) GetAdviseeStatistics(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	ctx := context.Background()

	// Get student IDs dari advisees
	studentIDs, err := s.studentRepo.FindStudentIDsByAdvisorID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data mahasiswa bimbingan",
		})
	}

	if len(studentIDs) == 0 {
		return c.Status(200).JSON(fiber.Map{
			"status":  "success",
			"message": "Tidak ada mahasiswa bimbingan",
			"data":    buildEmptyStatistics(true),
		})
	}

	// Get statistics dari MongoDB
	stats, err := s.achievementRepo.GetStatisticsByStudentIDs(ctx, studentIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil statistik",
		})
	}

	// Get top students
	topStudents, err := s.referenceRepo.GetTopStudents(studentIDs, 10)
	if err != nil {
		topStudents = []models.TopStudent{}
	}

	// Build response
	response := buildStatisticsResponse(stats, true)

	// Convert topStudents to fiber.Map
	topStudentsMap := []fiber.Map{}
	for _, student := range topStudents {
		topStudentsMap = append(topStudentsMap, fiber.Map{
			"student_id":            student.StudentID,
			"student_name":          student.StudentName,
			"total_achievements":    student.TotalAchievements,
			"verified_achievements": student.VerifiedAchievements,
		})
	}
	response["top_students"] = topStudentsMap

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Statistik mahasiswa bimbingan berhasil diambil",
		"data":    response,
	})
}

// GetAllStatistics godoc
// @Summary Get all achievement statistics
// @Description Admin view of comprehensive achievement statistics across all students including top performers ranking.
// @Tags Statistics & Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{status=string,message=string,data=object} "All statistics retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires admin access)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve statistics from database"
// @Router /reports/statistics [get]
func (s *AchievementService) GetAllStatistics(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get all student IDs (we'll use empty filter to get all)
	// For simplicity, we'll aggregate all achievements
	filter := bson.M{"is_deleted": false}
	achievements, err := s.achievementRepo.FindAll(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data achievements",
		})
	}

	// Calculate statistics manually
	stats := calculateStatisticsFromAchievements(achievements)

	// Get top students (all students)
	topStudents, err := s.referenceRepo.GetAllTopStudents(10)
	if err != nil {
		topStudents = []models.TopStudent{}
	}

	// Build response
	response := buildStatisticsResponse(stats, true)

	// Convert topStudents to fiber.Map
	topStudentsMap := []fiber.Map{}
	for _, student := range topStudents {
		topStudentsMap = append(topStudentsMap, fiber.Map{
			"student_id":            student.StudentID,
			"student_name":          student.StudentName,
			"total_achievements":    student.TotalAchievements,
			"verified_achievements": student.VerifiedAchievements,
		})
	}
	response["top_students"] = topStudentsMap

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Statistik semua prestasi berhasil diambil",
		"data":    response,
	})
}

// Helper function to build statistics response
func buildStatisticsResponse(stats map[string]interface{}, includeTopStudents bool) fiber.Map {
	totalAchievements := stats["total_achievements"].(int)

	// Summary
	summary := fiber.Map{
		"total_achievements": totalAchievements,
		"total_verified":     stats["total_verified"].(int),
		"total_pending":      stats["total_pending"].(int),
		"total_rejected":     stats["total_rejected"].(int),
		"total_draft":        stats["total_draft"].(int),
	}

	// By Category
	categoryCount := stats["category_count"].(map[string]int)
	byCategory := []fiber.Map{}
	for category, count := range categoryCount {
		percentage := 0.0
		if totalAchievements > 0 {
			percentage = float64(count) / float64(totalAchievements) * 100
		}
		byCategory = append(byCategory, fiber.Map{
			"category":   category,
			"count":      count,
			"percentage": percentage,
		})
	}

	// By Level
	levelCount := stats["level_count"].(map[string]int)
	byLevel := []fiber.Map{}
	for level, count := range levelCount {
		percentage := 0.0
		if totalAchievements > 0 {
			percentage = float64(count) / float64(totalAchievements) * 100
		}
		byLevel = append(byLevel, fiber.Map{
			"level":      level,
			"count":      count,
			"percentage": percentage,
		})
	}

	// By Period
	periodCount := stats["period_count"].(map[string]int)
	byPeriod := []fiber.Map{}
	for period, count := range periodCount {
		// Parse year-month
		var year, month int
		fmt.Sscanf(period, "%d-%d", &year, &month)
		byPeriod = append(byPeriod, fiber.Map{
			"year":  year,
			"month": month,
			"count": count,
		})
	}

	response := fiber.Map{
		"summary":     summary,
		"by_category": byCategory,
		"by_level":    byLevel,
		"by_period":   byPeriod,
	}

	if includeTopStudents {
		response["top_students"] = []fiber.Map{}
	}

	return response
}

// Helper function to build empty statistics
func buildEmptyStatistics(includeTopStudents bool) fiber.Map {
	response := fiber.Map{
		"summary": fiber.Map{
			"total_achievements": 0,
			"total_verified":     0,
			"total_pending":      0,
			"total_rejected":     0,
			"total_draft":        0,
		},
		"by_category": []fiber.Map{},
		"by_level":    []fiber.Map{},
		"by_period":   []fiber.Map{},
	}

	if includeTopStudents {
		response["top_students"] = []fiber.Map{}
	}

	return response
}

// Helper function to calculate statistics from achievements
func calculateStatisticsFromAchievements(achievements []models.Achievement) map[string]interface{} {
	stats := make(map[string]interface{})

	totalAchievements := len(achievements)
	totalVerified := 0
	totalPending := 0
	totalRejected := 0
	totalDraft := 0

	categoryCount := make(map[string]int)
	levelCount := make(map[string]int)
	periodCount := make(map[string]int)

	for _, achievement := range achievements {
		switch achievement.Status {
		case "verified":
			totalVerified++
		case "submitted":
			totalPending++
		case "rejected":
			totalRejected++
		case "draft":
			totalDraft++
		}

		if achievement.Category != "" {
			categoryCount[achievement.Category]++
		}

		if achievement.Level != "" {
			levelCount[achievement.Level]++
		}

		yearMonth := achievement.Date.Format("2006-01")
		periodCount[yearMonth]++
	}

	stats["total_achievements"] = totalAchievements
	stats["total_verified"] = totalVerified
	stats["total_pending"] = totalPending
	stats["total_rejected"] = totalRejected
	stats["total_draft"] = totalDraft
	stats["category_count"] = categoryCount
	stats["level_count"] = levelCount
	stats["period_count"] = periodCount

	return stats
}

// GetAchievementHistory godoc
// @Summary Get achievement history
// @Description Get status change history of an achievement including timestamps and verification/rejection details.
// @Tags Achievements
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Success 200 {object} object{status=string,message=string,data=object{achievement_id=string,current_status=string,history=[]object}} "History retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Access denied - not owner, admin, or lecturer"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve history data"
// @Router /achievements/{id}/history [get]
func (s *AchievementService) GetAchievementHistory(c *fiber.Ctx) error {
	achievementID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)

	ctx := context.Background()

	// Get achievement
	achievement, err := s.achievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Achievement tidak ditemukan",
		})
	}

	// Check access (owner or admin/lecturer)
	roleID, _ := c.Locals("role_id").(string)
	if achievement.StudentID != userID && roleID != "1" && roleID != "2" {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Anda tidak memiliki akses ke achievement ini",
		})
	}

	// Get reference for history
	reference, err := s.referenceRepo.FindByMongoID(achievementID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil history",
		})
	}

	// Build history timeline
	history := []fiber.Map{
		{
			"status":    "draft",
			"timestamp": achievement.CreatedAt,
			"note":      "Achievement created",
		},
	}

	if reference.SubmittedAt != nil {
		history = append(history, fiber.Map{
			"status":    "submitted",
			"timestamp": *reference.SubmittedAt,
			"note":      "Submitted for verification",
		})
	}

	if reference.VerifiedAt != nil && reference.VerifiedBy != nil {
		history = append(history, fiber.Map{
			"status":      "verified",
			"timestamp":   *reference.VerifiedAt,
			"verified_by": *reference.VerifiedBy,
			"note":        "Achievement verified",
		})
	}

	if reference.Status == "rejected" && reference.RejectionNote != nil {
		history = append(history, fiber.Map{
			"status":         "rejected",
			"timestamp":      reference.UpdatedAt,
			"rejection_note": *reference.RejectionNote,
			"note":           "Achievement rejected",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "History berhasil diambil",
		"data": fiber.Map{
			"achievement_id": achievementID,
			"current_status": achievement.Status,
			"history":        history,
		},
	})
}

// UploadAttachment godoc
// @Summary Upload additional attachments
// @Description Upload additional files to a draft achievement. Validates file types and handles rollback on errors.
// @Tags Achievements
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "Achievement ID"
// @Param attachments formData file true "Additional attachment files (multiple files allowed)"
// @Success 200 {object} object{status=string,message=string,data=object{achievement_id=string,new_documents=[]models.Document,total_documents=int}} "Attachments uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request, no files, or upload error"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Access denied or achievement not in draft status"
// @Failure 404 {object} map[string]interface{} "Achievement not found"
// @Failure 500 {object} map[string]interface{} "Upload operation failed"
// @Router /achievements/{id}/attachments [post]
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	achievementID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)

	ctx := context.Background()

	// Get achievement
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

	// Check status (only draft can add attachments)
	if achievement.Status != "draft" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Hanya achievement dengan status 'draft' yang bisa menambah attachment",
		})
	}

	// Handle file upload
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid form data",
		})
	}

	files := form.File["attachments"]
	if len(files) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "No files uploaded",
		})
	}

	var newDocuments []models.Document
	for _, file := range files {
		// Save file
		filepath, err := utils.SaveUploadedFile(file, s.uploadConfig)
		if err != nil {
			// Rollback uploaded files
			for _, doc := range newDocuments {
				utils.DeleteFile(doc.Filepath)
			}
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Gagal upload file: %v", err),
			})
		}

		// Add to documents
		newDocuments = append(newDocuments, models.Document{
			Filename:   file.Filename,
			Filepath:   filepath,
			Filesize:   file.Size,
			Mimetype:   file.Header.Get("Content-Type"),
			UploadedAt: time.Now(),
		})
	}

	// Append to existing documents
	achievement.Documents = append(achievement.Documents, newDocuments...)
	achievement.UpdatedAt = time.Now()

	// Update in MongoDB
	if err := s.achievementRepo.Update(ctx, achievementID, achievement); err != nil {
		// Rollback uploaded files
		for _, doc := range newDocuments {
			utils.DeleteFile(doc.Filepath)
		}
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate achievement",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Attachment berhasil diupload",
		"data": fiber.Map{
			"achievement_id":  achievementID,
			"new_documents":   newDocuments,
			"total_documents": len(achievement.Documents),
		},
	})
}

// GetStudentAchievements godoc
// @Summary Get student achievements
// @Description Get all achievements for a specific student. Access control: admin, lecturer, or the student themselves.
// @Tags Student Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Success 200 {object} object{status=string,message=string,data=object{student_id=string,achievements=[]models.Achievement,total=int}} "Student achievements retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Access denied - insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve achievements"
// @Router /students/{id}/achievements [get]
func (s *AchievementService) GetStudentAchievements(c *fiber.Ctx) error {
	studentID := c.Params("id")

	// Check access
	userID, _ := c.Locals("user_id").(string)
	roleID, _ := c.Locals("role_id").(string)

	// Only admin, lecturer, or the student themselves can view
	if userID != studentID && roleID != "1" && roleID != "2" {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Anda tidak memiliki akses ke data ini",
		})
	}

	ctx := context.Background()
	achievements, err := s.achievementRepo.FindByStudentID(ctx, studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data achievements",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data achievements berhasil diambil",
		"data": fiber.Map{
			"student_id":   studentID,
			"achievements": achievements,
			"total":        len(achievements),
		},
	})
}

// GetStudentReport godoc
// @Summary Get student report
// @Description Get detailed achievement report for a specific student including statistics and achievement list.
// @Tags Statistics & Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Success 200 {object} object{status=string,message=string,data=object{student=models.Student,statistics=object,achievements=[]models.Achievement}} "Student report retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing JWT token"
// @Failure 403 {object} map[string]interface{} "Access denied - not owner, admin, or lecturer"
// @Failure 404 {object} map[string]interface{} "Student not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve report data"
// @Router /reports/student/{id} [get]
func (s *AchievementService) GetStudentReport(c *fiber.Ctx) error {
	studentID := c.Params("id")

	// Check access
	userID, _ := c.Locals("user_id").(string)
	roleID, _ := c.Locals("role_id").(string)

	// Only admin, lecturer, or the student themselves can view
	if userID != studentID && roleID != "1" && roleID != "2" {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Anda tidak memiliki akses ke data ini",
		})
	}

	ctx := context.Background()

	// Get student info
	student, err := s.studentRepo.FindByUserID(studentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Student tidak ditemukan",
		})
	}

	// Get achievements
	achievements, err := s.achievementRepo.FindByStudentID(ctx, studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data achievements",
		})
	}

	// Calculate statistics
	stats := calculateStatisticsFromAchievements(achievements)

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Report berhasil diambil",
		"data": fiber.Map{
			"student":      student,
			"statistics":   buildStatisticsResponse(stats, false),
			"achievements": achievements,
		},
	})
}