package test

import (
	models "crud-app/app/model"
	"crud-app/app/utils"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key")

	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
		RoleID:   "1",
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	// Verify token structure
	if len(token) < 10 {
		t.Error("GenerateToken() returned token too short")
	}
}

func TestValidateToken(t *testing.T) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key")

	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
		RoleID:   "1",
	}

	// Generate valid token
	validToken, err := utils.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "Invalid token",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Malformed token",
			token:   "not.a.jwt",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := utils.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims == nil {
					t.Error("ValidateToken() returned nil claims for valid token")
				}
				if claims.UserID != user.ID {
					t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, user.ID)
				}
				if claims.Username != user.Username {
					t.Errorf("ValidateToken() Username = %v, want %v", claims.Username, user.Username)
				}
				if claims.RoleID != user.RoleID {
					t.Errorf("ValidateToken() RoleID = %v, want %v", claims.RoleID, user.RoleID)
				}
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	// Set test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key")

	// Create expired token manually
	claims := utils.JwtClaims{
		UserID:   "test-user",
		Username: "testuser",
		RoleID:   "1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	// Try to validate expired token
	_, err = utils.ValidateToken(tokenString)
	if err == nil {
		t.Error("ValidateToken() should return error for expired token")
	}
}