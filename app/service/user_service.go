package service

import (
	models "crud-app/app/model"
	"crud-app/app/repository"
	"crud-app/app/utils"
	"crypto/rand"
	"database/sql"
	"math/big"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo     *repository.UserRepository
	studentRepo  *repository.StudentRepository
	lecturerRepo *repository.LecturerRepository
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		userRepo:     repository.NewUserRepository(db),
		studentRepo:  repository.NewStudentRepository(db),
		lecturerRepo: repository.NewLecturerRepository(db),
	}
}

// CreateUser - FR-009: Create new user
func (s *UserService) CreateUser(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
		IsActive bool   `json:"is_active"`
		// Student fields (optional)
		StudentID    string `json:"student_id"`
		ProgramStudy string `json:"program_study"`
		AcademicYear string `json:"academic_year"`
		// Lecturer fields (optional)
		LecturerID string `json:"lecturer_id"`
		Department string `json:"department"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validation
	if req.Username == "" || req.Email == "" || req.FullName == "" || req.RoleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Username, email, full_name, dan role_id harus diisi",
		})
	}

	// Check username exists
	exists, err := s.userRepo.CheckUsernameExists(req.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengecek username",
		})
	}
	if exists {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Username sudah digunakan",
		})
	}

	// Check email exists
	exists, err = s.userRepo.CheckEmailExists(req.Email)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengecek email",
		})
	}
	if exists {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Email sudah digunakan",
		})
	}

	// Check role exists
	roleExists, err := s.userRepo.CheckRoleExists(req.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengecek role",
		})
	}
	if !roleExists {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Role tidak ditemukan",
		})
	}

	// Generate random password
	plainPassword := generateRandomPassword(12)
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal generate password",
		})
	}

	// Create user
	userID := uuid.New().String()
	user := &models.User{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       req.RoleID,
		IsActive:     req.IsActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal membuat user",
		})
	}

	// Create student profile if role is student (role_id = "3")
	if req.RoleID == "3" && req.StudentID != "" {
		student := &models.Student{
			ID:           uuid.New().String(),
			UserID:       userID,
			StudentID:    req.StudentID,
			ProgramStudy: req.ProgramStudy,
			AcademicYear: req.AcademicYear,
			CreatedAt:    time.Now(),
		}
		if err := s.studentRepo.Create(student); err != nil {
			// Rollback user creation
			s.userRepo.SoftDelete(userID)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Gagal membuat student profile",
			})
		}
	}

	// Create lecturer profile if role is lecturer (role_id = "2")
	if req.RoleID == "2" && req.LecturerID != "" {
		lecturer := &models.Lecturer{
			ID:         uuid.New().String(),
			UserID:     userID,
			LecturerID: req.LecturerID,
			Department: req.Department,
			CreatedAt:  time.Now(),
		}
		if err := s.lecturerRepo.Create(lecturer); err != nil {
			// Rollback user creation
			s.userRepo.SoftDelete(userID)
			return c.Status(500).JSON(fiber.Map{
				"status":  "error",
				"message": "Gagal membuat lecturer profile",
			})
		}
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "User berhasil dibuat",
		"data": fiber.Map{
			"user_id":  userID,
			"username": req.Username,
			"email":    req.Email,
			"password": plainPassword, // Return plain password untuk diberikan ke user
			"role_id":  req.RoleID,
		},
	})
}

// GetUsers - FR-009: Get list users dengan pagination
func (s *UserService) GetUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	roleFilter := c.Query("role_id", "")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	users, total, err := s.userRepo.FindAll(limit, offset, roleFilter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data users",
		})
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data users berhasil diambil",
		"data": fiber.Map{
			"users": users,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total_items": total,
				"total_pages": totalPages,
			},
		},
	})
}

// GetUserByID - FR-009: Get user detail
func (s *UserService) GetUserByID(c *fiber.Ctx) error {
	userID := c.Params("id")

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak ditemukan",
		})
	}

	// Get profile based on role
	var profile interface{}
	if user.RoleID == "3" { // Student
		profile, _ = s.studentRepo.FindByUserID(userID)
	} else if user.RoleID == "2" { // Lecturer
		profile, _ = s.lecturerRepo.FindByUserID(userID)
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data user berhasil diambil",
		"data": fiber.Map{
			"user":    user,
			"profile": profile,
		},
	})
}

// UpdateUser - FR-009: Update user
func (s *UserService) UpdateUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Get existing user
	existing, err := s.userRepo.FindByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak ditemukan",
		})
	}

	// Update fields
	if req.Username != "" {
		existing.Username = req.Username
	}
	if req.Email != "" {
		existing.Email = req.Email
	}
	if req.FullName != "" {
		existing.FullName = req.FullName
	}
	if req.RoleID != "" {
		existing.RoleID = req.RoleID
	}
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()

	if err := s.userRepo.Update(userID, existing); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate user",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "User berhasil diupdate",
		"data":    existing,
	})
}

// DeleteUser - FR-009: Soft delete user
func (s *UserService) DeleteUser(c *fiber.Ctx) error {
	userID := c.Params("id")

	// Check if user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak ditemukan",
		})
	}

	// Soft delete user
	if err := s.userRepo.SoftDelete(userID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menghapus user",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "User berhasil dihapus",
	})
}

// AssignRole - FR-009: Assign role to user
func (s *UserService) AssignRole(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req struct {
		RoleID string `json:"role_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if req.RoleID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Role ID harus diisi",
		})
	}

	// Check role exists
	roleExists, err := s.userRepo.CheckRoleExists(req.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengecek role",
		})
	}
	if !roleExists {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Role tidak ditemukan",
		})
	}

	// Assign role
	if err := s.userRepo.AssignRole(userID, req.RoleID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal assign role",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Role berhasil diassign",
	})
}

// SetStudentProfile - FR-009: Set student profile
func (s *UserService) SetStudentProfile(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req struct {
		StudentID    string `json:"student_id"`
		ProgramStudy string `json:"program_study"`
		AcademicYear string `json:"academic_year"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validation
	if req.StudentID == "" || req.ProgramStudy == "" || req.AcademicYear == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Student ID, program study, dan academic year harus diisi",
		})
	}

	// Check if student profile already exists
	existing, _ := s.studentRepo.FindByUserID(userID)
	if existing != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Student profile sudah ada. Gunakan endpoint update",
		})
	}

	// Create student profile
	student := &models.Student{
		ID:           uuid.New().String(),
		UserID:       userID,
		StudentID:    req.StudentID,
		ProgramStudy: req.ProgramStudy,
		AcademicYear: req.AcademicYear,
		CreatedAt:    time.Now(),
	}

	if err := s.studentRepo.Create(student); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal membuat student profile",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Student profile berhasil dibuat",
		"data":    student,
	})
}

// UpdateStudentProfile - FR-009: Update student profile
func (s *UserService) UpdateStudentProfile(c *fiber.Ctx) error {
	studentID := c.Params("id")

	var req struct {
		StudentID    string `json:"student_id"`
		ProgramStudy string `json:"program_study"`
		AcademicYear string `json:"academic_year"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Get existing student
	existing, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Student tidak ditemukan",
		})
	}

	// Update
	student := &models.Student{
		StudentID:    req.StudentID,
		ProgramStudy: req.ProgramStudy,
		AcademicYear: req.AcademicYear,
		AdvisorID:    existing.AdvisorID,
	}

	if err := s.studentRepo.Update(studentID, student); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate student profile",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Student profile berhasil diupdate",
	})
}

// AssignAdvisor - FR-009: Assign advisor to student
func (s *UserService) AssignAdvisor(c *fiber.Ctx) error {
	studentID := c.Params("id")

	var req struct {
		AdvisorID string `json:"advisor_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if req.AdvisorID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Advisor ID harus diisi",
		})
	}

	// Check if advisor is a lecturer
	lecturerExists, err := s.lecturerRepo.CheckExists(req.AdvisorID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengecek lecturer",
		})
	}
	if !lecturerExists {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Advisor harus seorang lecturer",
		})
	}

	// Assign advisor
	if err := s.studentRepo.AssignAdvisor(studentID, req.AdvisorID); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal assign advisor",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Advisor berhasil diassign",
	})
}

// SetLecturerProfile - FR-009: Set lecturer profile
func (s *UserService) SetLecturerProfile(c *fiber.Ctx) error {
	userID := c.Params("id")

	var req struct {
		LecturerID string `json:"lecturer_id"`
		Department string `json:"department"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validation
	if req.LecturerID == "" || req.Department == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Lecturer ID dan department harus diisi",
		})
	}

	// Check if lecturer profile already exists
	existing, _ := s.lecturerRepo.FindByUserID(userID)
	if existing != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Lecturer profile sudah ada. Gunakan endpoint update",
		})
	}

	// Create lecturer profile
	lecturer := &models.Lecturer{
		ID:         uuid.New().String(),
		UserID:     userID,
		LecturerID: req.LecturerID,
		Department: req.Department,
		CreatedAt:  time.Now(),
	}

	if err := s.lecturerRepo.Create(lecturer); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal membuat lecturer profile",
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Lecturer profile berhasil dibuat",
		"data":    lecturer,
	})
}

// UpdateLecturerProfile - FR-009: Update lecturer profile
func (s *UserService) UpdateLecturerProfile(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	var req struct {
		LecturerID string `json:"lecturer_id"`
		Department string `json:"department"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Update
	lecturer := &models.Lecturer{
		LecturerID: req.LecturerID,
		Department: req.Department,
	}

	if err := s.lecturerRepo.Update(lecturerID, lecturer); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengupdate lecturer profile",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Lecturer profile berhasil diupdate",
	})
}

// Helper function to generate random password
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
	password := make([]byte, length)
	for i := range password {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[num.Int64()]
	}
	return string(password)
}