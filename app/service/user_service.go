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

// CreateUser godoc
// @Summary Create new user
// @Description Create a new user with role assignment and optional student/lecturer profile. Generates random password.
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{username=string,email=string,full_name=string,role_id=string,is_active=bool,student_id=string,program_study=string,academic_year=string,lecturer_id=string,department=string} true "User creation request"
// @Success 201 {object} object{status=string,message=string,data=object{user_id=string,username=string,email=string,password=string,role_id=string}} "User created successfully with generated password"
// @Failure 400 {object} map[string]interface{} "Invalid request, validation error, or duplicate username/email"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires users.create)"
// @Failure 500 {object} map[string]interface{} "Internal server error - user or profile creation failed"
// @Router /users [post]
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

// GetUsers godoc
// @Summary Get list of users
// @Description Get paginated list of users with optional role filtering. Admin access required.
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" default(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" default(10)
// @Param role_id query string false "Filter by role ID (1=Admin, 2=Lecturer, 3=Student)"
// @Success 200 {object} object{status=string,message=string,data=object{users=[]models.User,pagination=models.PaginationMeta}} "Users retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires users.read)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve users"
// @Router /users [get]
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

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get detailed information about a specific user including role-based profile (Student/Lecturer).
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} object{status=string,message=string,data=object{user=models.User,profile=object}} "User retrieved successfully with profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires users.read)"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve user data"
// @Router /users/{id} [get]
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

// UpdateUser godoc
// @Summary Update user
// @Description Update user information including role change and account activation/deactivation.
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body object{username=string,email=string,full_name=string,role_id=string,is_active=bool} true "User update request"
// @Success 200 {object} object{status=string,message=string,data=models.User} "User updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires users.update)"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Update operation failed"
// @Router /users/{id} [put]
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

// DeleteUser godoc
// @Summary Delete user
// @Description Soft delete a user. User data is marked as deleted but not physically removed.
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} object{status=string,message=string} "User deleted successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires users.delete)"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Delete operation failed"
// @Router /users/{id} [delete]
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

// AssignRole godoc
// @Summary Assign role to user
// @Description Assign a specific role to a user. Validates role existence before assignment.
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Param request body object{role_id=string} true "Role assignment request (1=Admin, 2=Lecturer, 3=Student)"
// @Success 200 {object} object{status=string,message=string} "Role assigned successfully"
// @Failure 400 {object} map[string]interface{} "Invalid role ID or role not found"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires users.assign_role)"
// @Failure 500 {object} map[string]interface{} "Role assignment failed"
// @Router /users/{id}/role [put]
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

// SetStudentProfile godoc
// @Summary Set student profile
// @Description Create student profile for a user. Validates that profile doesn't already exist.
// @Tags Student Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body object{student_id=string,program_study=string,academic_year=string} true "Student profile creation request"
// @Success 201 {object} object{status=string,message=string,data=models.Student} "Student profile created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or profile already exists"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Profile creation failed"
// @Router /users/{id}/student-profile [post]
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

// UpdateStudentProfile godoc
// @Summary Update student profile
// @Description Update student profile information including student ID, program study, and academic year.
// @Tags Student Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Param request body object{student_id=string,program_study=string,academic_year=string} true "Student profile update request"
// @Success 200 {object} object{status=string,message=string} "Student profile updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Student not found"
// @Failure 500 {object} map[string]interface{} "Update operation failed"
// @Router /students/{id}/profile [put]
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

// AssignAdvisor godoc
// @Summary Assign advisor to student
// @Description Assign a lecturer as advisor to a student. Validates that the advisor is a valid lecturer.
// @Tags Student Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Param request body object{advisor_id=string} true "Advisor assignment request with lecturer ID"
// @Success 200 {object} object{status=string,message=string} "Advisor assigned successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or advisor must be a lecturer"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires students.assign_advisor)"
// @Failure 500 {object} map[string]interface{} "Advisor assignment failed"
// @Router /students/{id}/advisor [put]
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

// SetLecturerProfile godoc
// @Summary Set lecturer profile
// @Description Create lecturer profile for a user. Validates that profile doesn't already exist.
// @Tags Lecturer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body object{lecturer_id=string,department=string} true "Lecturer profile creation request"
// @Success 201 {object} object{status=string,message=string,data=models.Lecturer} "Lecturer profile created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or profile already exists"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions"
// @Failure 500 {object} map[string]interface{} "Profile creation failed"
// @Router /users/{id}/lecturer-profile [post]
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

// UpdateLecturerProfile godoc
// @Summary Update lecturer profile
// @Description Update lecturer profile information including lecturer ID and department.
// @Tags Lecturer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lecturer ID"
// @Param request body object{lecturer_id=string,department=string} true "Lecturer profile update request"
// @Success 200 {object} object{status=string,message=string} "Lecturer profile updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions"
// @Failure 404 {object} map[string]interface{} "Lecturer not found"
// @Failure 500 {object} map[string]interface{} "Update operation failed"
// @Router /lecturers/{id}/profile [put]
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

// GetStudents godoc
// @Summary Get list of students
// @Description Get paginated list of students with their profile information.
// @Tags Student Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" default(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" default(10)
// @Success 200 {object} object{status=string,message=string,data=object{students=[]models.Student,pagination=models.PaginationMeta}} "Students retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires students.read)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve students"
// @Router /students [get]
func (s *UserService) GetStudents(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	students, total, err := s.studentRepo.FindAll(limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data students",
		})
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data students berhasil diambil",
		"data": fiber.Map{
			"students": students,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total_items": total,
				"total_pages": totalPages,
			},
		},
	})
}

// GetStudentByID godoc
// @Summary Get student by ID
// @Description Get detailed information about a specific student including user data and profile.
// @Tags Student Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Success 200 {object} object{status=string,message=string,data=object{student=models.Student,user=models.User}} "Student retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires students.read)"
// @Failure 404 {object} map[string]interface{} "Student not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve student data"
// @Router /students/{id} [get]
func (s *UserService) GetStudentByID(c *fiber.Ctx) error {
	studentID := c.Params("id")

	student, err := s.studentRepo.FindByID(studentID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Student tidak ditemukan",
		})
	}

	// Get user data
	user, err := s.userRepo.FindByID(student.UserID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data user",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data student berhasil diambil",
		"data": fiber.Map{
			"student": student,
			"user":    user,
		},
	})
}

// GetLecturers godoc
// @Summary Get list of lecturers
// @Description Get paginated list of lecturers with their profile information and department details.
// @Tags Lecturer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" default(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" default(10)
// @Success 200 {object} object{status=string,message=string,data=object{lecturers=[]models.Lecturer,pagination=models.PaginationMeta}} "Lecturers retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires lecturers.read)"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve lecturers"
// @Router /lecturers [get]
func (s *UserService) GetLecturers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	lecturers, total, err := s.lecturerRepo.FindAll(limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data lecturers",
		})
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data lecturers berhasil diambil",
		"data": fiber.Map{
			"lecturers": lecturers,
			"pagination": fiber.Map{
				"page":        page,
				"limit":       limit,
				"total_items": total,
				"total_pages": totalPages,
			},
		},
	})
}

// GetAdvisees godoc
// @Summary Get lecturer advisees
// @Description Get list of students advised by a specific lecturer including their profile information.
// @Tags Lecturer Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lecturer ID"
// @Success 200 {object} object{status=string,message=string,data=object{advisees=[]models.Student,total=int}} "Advisees retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Insufficient permissions (requires lecturers.read)"
// @Failure 404 {object} map[string]interface{} "Lecturer not found"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve advisees"
// @Router /lecturers/{id}/advisees [get]
func (s *UserService) GetAdvisees(c *fiber.Ctx) error {
	lecturerID := c.Params("id")

	// Check if lecturer exists
	_, err := s.lecturerRepo.FindByID(lecturerID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Lecturer tidak ditemukan",
		})
	}

	// Get advisees
	advisees, err := s.studentRepo.FindByAdvisorID(lecturerID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data advisees",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Data advisees berhasil diambil",
		"data": fiber.Map{
			"advisees": advisees,
			"total":    len(advisees),
		},
	})
}