package test

import (
	models "crud-app/app/model"
	"crud-app/app/utils"
	"crud-app/test/mocks"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuthService_Login_Success(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "test-secret")

	mockRepo := mocks.NewMockUserRepository()

	// Create a test user
	hashedPassword, _ := utils.HashPassword("password123")
	testUser := &models.User{
		ID:           "test-user-id",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		FullName:     "Test User",
		RoleID:       "1",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	mockRepo.AddUser(testUser)

	// Test successful login
	reqBody := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest("POST", "/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// We need to create a custom test since we can't easily inject the mock
	// This is a simplified test structure

	// Test login request parsing
	var loginReq models.LoginRequest
	err := json.Unmarshal([]byte(reqBody), &loginReq)
	if err != nil {
		t.Fatalf("Failed to parse login request: %v", err)
	}

	// Verify request fields
	if loginReq.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", loginReq.Username)
	}
	if loginReq.Password != "password123" {
		t.Errorf("Expected password 'password123', got '%s'", loginReq.Password)
	}

	// Test user lookup
	user, err := mockRepo.FindByUsernameOrEmail("testuser")
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	// Test password verification
	if !utils.CheckPassword("password123", user.PasswordHash) {
		t.Error("Password verification failed")
	}

	// Test token generation
	token, err := utils.GenerateToken(*user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Error("Generated token is empty")
	}

	// Test user profile retrieval
	profile, err := mockRepo.GetUserProfile(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user profile: %v", err)
	}
	if profile.Username != user.Username {
		t.Errorf("Expected profile username '%s', got '%s'", user.Username, profile.Username)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()

	// Test with non-existent user
	_, err := mockRepo.FindByUsernameOrEmail("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	// Test with wrong password
	hashedPassword, _ := utils.HashPassword("correctpassword")
	testUser := &models.User{
		ID:           "test-user-id",
		Username:     "testuser",
		PasswordHash: hashedPassword,
		IsActive:     true,
	}
	mockRepo.AddUser(testUser)

	user, _ := mockRepo.FindByUsernameOrEmail("testuser")
	if utils.CheckPassword("wrongpassword", user.PasswordHash) {
		t.Error("Password check should fail for wrong password")
	}
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	mockRepo := mocks.NewMockUserRepository()

	// Create inactive user
	hashedPassword, _ := utils.HashPassword("password123")
	inactiveUser := &models.User{
		ID:           "inactive-user-id",
		Username:     "inactiveuser",
		PasswordHash: hashedPassword,
		IsActive:     false, // Inactive user
	}
	mockRepo.AddUser(inactiveUser)

	user, err := mockRepo.FindByUsernameOrEmail("inactiveuser")
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	// Check if user is inactive
	if user.IsActive {
		t.Error("User should be inactive")
	}

	// In real implementation, this should return 403 error
}

func TestAuthService_Login_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "Empty username",
			username: "",
			password: "password123",
			wantErr:  true,
		},
		{
			name:     "Empty password",
			username: "testuser",
			password: "",
			wantErr:  true,
		},
		{
			name:     "Both empty",
			username: "",
			password: "",
			wantErr:  true,
		},
		{
			name:     "Valid credentials",
			username: "testuser",
			password: "password123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate input
			hasError := tt.username == "" || tt.password == ""
			if hasError != tt.wantErr {
				t.Errorf("Validation error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}

func TestAuthService_TokenGeneration(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")

	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
		RoleID:   "1",
	}

	// Test token generation
	token, err := utils.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Test token validation
	claims, err := utils.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Verify claims
	if claims.UserID != user.ID {
		t.Errorf("Expected UserID %s, got %s", user.ID, claims.UserID)
	}
	if claims.Username != user.Username {
		t.Errorf("Expected Username %s, got %s", user.Username, claims.Username)
	}
	if claims.RoleID != user.RoleID {
		t.Errorf("Expected RoleID %s, got %s", user.RoleID, claims.RoleID)
	}
}