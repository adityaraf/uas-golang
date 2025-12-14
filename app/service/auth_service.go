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

// Login godoc
// @Summary User login
// @Description Authenticate user with username/email and password. Returns JWT token and user profile information.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials (username/email and password)"
// @Success 200 {object} models.LoginResponse "Login successful - returns JWT token and user profile"
// @Failure 400 {object} map[string]interface{} "Invalid request body or missing required fields"
// @Failure 401 {object} map[string]interface{} "Invalid credentials (wrong username/email or password)"
// @Failure 403 {object} map[string]interface{} "Account inactive - contact administrator"
// @Failure 500 {object} map[string]interface{} "Internal server error - token generation failed"
// @Router /auth/login [post]
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

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Generate new access token using refresh token. Validates refresh token and user status.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token request"
// @Success 200 {object} object{status=string,message=string,data=object{token=string}} "Token refreshed successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or missing refresh token"
// @Failure 401 {object} map[string]interface{} "Invalid or expired refresh token"
// @Failure 403 {object} map[string]interface{} "User account inactive"
// @Failure 500 {object} map[string]interface{} "Failed to generate new token"
// @Router /auth/refresh [post]
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

// Logout godoc
// @Summary User logout
// @Description Logout user (client-side token removal). JWT tokens are stateless, so logout is handled client-side.
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{status=string,message=string} "Logout successful"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing token"
// @Router /auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
	// In JWT, logout is typically handled client-side by removing the token
	// Here we just return success message
	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Logout berhasil",
	})
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get profile information of the currently authenticated user including role and status.
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{status=string,message=string,data=models.UserProfile} "Profile retrieved successfully"
// @Failure 401 {object} map[string]interface{} "Unauthorized - invalid or missing token"
// @Failure 500 {object} map[string]interface{} "Failed to retrieve profile data"
// @Router /auth/profile [get]
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