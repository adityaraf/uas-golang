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

// RefreshToken - Refresh JWT token
func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Refresh token harus diisi",
		})
	}

	// Validate refresh token
	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid refresh token",
		})
	}

	// Get user from database
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak ditemukan",
		})
	}

	// Check if user is active
	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{
			"status":  "error",
			"message": "Akun Anda tidak aktif",
		})
	}

	// Generate new token
	newToken, err := utils.GenerateToken(*user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal generate token",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Token berhasil direfresh",
		"data": fiber.Map{
			"token": newToken,
		},
	})
}

// Logout - Logout user (client-side token removal)
func (s *AuthService) Logout(c *fiber.Ctx) error {
	// In JWT, logout is typically handled client-side by removing the token
	// Here we just return success message
	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Logout berhasil",
	})
}

// GetProfile - Get current user profile
func (s *AuthService) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Unauthorized",
		})
	}

	// Get user profile
	profile, err := s.userRepo.GetUserProfile(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data profile",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Profile berhasil diambil",
		"data":    profile,
	})
}