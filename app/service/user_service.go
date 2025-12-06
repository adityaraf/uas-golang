package service

import (
	models "crud-app/app/model"
	"crud-app/app/repository"
	"crud-app/app/utils"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(db),
	}
}

// Login service
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req models.LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	// Validasi input
	if req.Username == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Username dan password harus diisi",
		})
	}

	// Cari user berdasarkan username atau email
	user, err := s.userRepo.FindByUsernameOrEmail(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Username or Email salah",
		})
	}

	// Cek status aktif user
	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Akun Anda tidak aktif. Silakan hubungi administrator",
		})
	}

	// Validasi password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "password salah",
		})
	}

	// Generate JWT token
	token, err := utils.GenerateToken(*user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal generate token",
		})
	}

	// Get user profile dengan role name
	profile, err := s.userRepo.GetUserProfile(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data profile",
		})
	}

	// Return response
	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Login berhasil",
		"data": fiber.Map{
			"token":   token,
			"profile": profile,
		},
	})
}