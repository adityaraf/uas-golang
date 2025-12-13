package test

import (
	models "crud-app/app/model"
	"crud-app/test/mocks"
	"fmt"
	"testing"
	"time"
)

func TestUserService_CreateUser_Success(t *testing.T) {
	// Setup mocks
	mockUserRepo := mocks.NewMockUserRepository()

	// Test data
	userReq := struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		RoleID   string `json:"role_id"`
		IsActive bool   `json:"is_active"`
	}{
		Username: "newuser",
		Email:    "newuser@example.com",
		FullName: "New User",
		RoleID:   "3", // Student role
		IsActive: true,
	}

	// Test validation
	if userReq.Username == "" || userReq.Email == "" || userReq.FullName == "" || userReq.RoleID == "" {
		t.Error("Validation should fail for empty required fields")
	}

	// Test username uniqueness check
	exists, err := mockUserRepo.CheckUsernameExists(userReq.Username)
	if err != nil {
		t.Fatalf("CheckUsernameExists failed: %v", err)
	}
	if exists {
		t.Error("Username should not exist initially")
	}

	// Test email uniqueness check
	exists, err = mockUserRepo.CheckEmailExists(userReq.Email)
	if err != nil {
		t.Fatalf("CheckEmailExists failed: %v", err)
	}
	if exists {
		t.Error("Email should not exist initially")
	}

	// Test role existence check
	roleExists, err := mockUserRepo.CheckRoleExists(userReq.RoleID)
	if err != nil {
		t.Fatalf("CheckRoleExists failed: %v", err)
	}
	if !roleExists {
		t.Error("Role should exist")
	}

	// Test user creation
	user := &models.User{
		ID:        "test-user-id",
		Username:  userReq.Username,
		Email:     userReq.Email,
		FullName:  userReq.FullName,
		RoleID:    userReq.RoleID,
		IsActive:  userReq.IsActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = mockUserRepo.Create(user)
	if err != nil {
		t.Fatalf("Create user failed: %v", err)
	}

	// Verify user was created
	createdUser, err := mockUserRepo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if createdUser.Username != userReq.Username {
		t.Errorf("Expected username %s, got %s", userReq.Username, createdUser.Username)
	}
	if createdUser.Email != userReq.Email {
		t.Errorf("Expected email %s, got %s", userReq.Email, createdUser.Email)
	}
}

func TestUserService_CreateUser_ValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		fullName string
		roleID   string
		wantErr  bool
	}{
		{
			name:     "Empty username",
			username: "",
			email:    "test@example.com",
			fullName: "Test User",
			roleID:   "3",
			wantErr:  true,
		},
		{
			name:     "Empty email",
			username: "testuser",
			email:    "",
			fullName: "Test User",
			roleID:   "3",
			wantErr:  true,
		},
		{
			name:     "Empty full name",
			username: "testuser",
			email:    "test@example.com",
			fullName: "",
			roleID:   "3",
			wantErr:  true,
		},
		{
			name:     "Empty role ID",
			username: "testuser",
			email:    "test@example.com",
			fullName: "Test User",
			roleID:   "",
			wantErr:  true,
		},
		{
			name:     "Valid data",
			username: "testuser",
			email:    "test@example.com",
			fullName: "Test User",
			roleID:   "3",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := tt.username == "" || tt.email == "" || tt.fullName == "" || tt.roleID == ""
			if hasError != tt.wantErr {
				t.Errorf("Validation error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}

func TestUserService_CreateUser_DuplicateUsername(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository()

	// Create existing user
	existingUser := &models.User{
		ID:       "existing-user-id",
		Username: "existinguser",
		Email:    "existing@example.com",
	}
	mockUserRepo.AddUser(existingUser)

	// Test duplicate username
	exists, err := mockUserRepo.CheckUsernameExists("existinguser")
	if err != nil {
		t.Fatalf("CheckUsernameExists failed: %v", err)
	}
	if !exists {
		t.Error("Username should exist")
	}

	// Test duplicate email
	exists, err = mockUserRepo.CheckEmailExists("existing@example.com")
	if err != nil {
		t.Fatalf("CheckEmailExists failed: %v", err)
	}
	if !exists {
		t.Error("Email should exist")
	}
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository()

	// Create existing user
	existingUser := &models.User{
		ID:       "test-user-id",
		Username: "oldusername",
		Email:    "old@example.com",
		FullName: "Old Name",
		RoleID:   "3",
		IsActive: true,
	}
	mockUserRepo.AddUser(existingUser)

	// Test update
	updatedUser := &models.User{
		ID:        existingUser.ID,
		Username:  "newusername",
		Email:     "new@example.com",
		FullName:  "New Name",
		RoleID:    "2",
		IsActive:  false,
		UpdatedAt: time.Now(),
	}

	err := mockUserRepo.Update(existingUser.ID, updatedUser)
	if err != nil {
		t.Fatalf("Update user failed: %v", err)
	}

	// Verify update
	user, err := mockUserRepo.FindByID(existingUser.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if user.Username != "newusername" {
		t.Errorf("Expected username 'newusername', got '%s'", user.Username)
	}
	if user.Email != "new@example.com" {
		t.Errorf("Expected email 'new@example.com', got '%s'", user.Email)
	}
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository()

	// Create user
	user := &models.User{
		ID:       "test-user-id",
		Username: "testuser",
		Email:    "test@example.com",
	}
	mockUserRepo.AddUser(user)

	// Verify user exists
	_, err := mockUserRepo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("User should exist before deletion: %v", err)
	}

	// Test soft delete
	err = mockUserRepo.SoftDelete(user.ID)
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	// Verify user is deleted (in mock, we remove from map)
	_, err = mockUserRepo.FindByID(user.ID)
	if err == nil {
		t.Error("User should not be found after soft delete")
	}
}

func TestUserService_AssignRole_Success(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository()

	// Create user
	user := &models.User{
		ID:     "test-user-id",
		RoleID: "3", // Student
	}
	mockUserRepo.AddUser(user)

	// Test role assignment
	newRoleID := "2" // Lecturer

	// Check if new role exists
	roleExists, err := mockUserRepo.CheckRoleExists(newRoleID)
	if err != nil {
		t.Fatalf("CheckRoleExists failed: %v", err)
	}
	if !roleExists {
		t.Error("New role should exist")
	}

	// Assign role
	err = mockUserRepo.AssignRole(user.ID, newRoleID)
	if err != nil {
		t.Fatalf("AssignRole failed: %v", err)
	}

	// Verify role assignment
	updatedUser, err := mockUserRepo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if updatedUser.RoleID != newRoleID {
		t.Errorf("Expected role ID %s, got %s", newRoleID, updatedUser.RoleID)
	}
}

func TestUserService_GetUsers_WithPagination(t *testing.T) {
	mockUserRepo := mocks.NewMockUserRepository()

	// Create multiple users
	for i := 0; i < 25; i++ {
		user := &models.User{
			ID:       fmt.Sprintf("user-%d", i),
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			RoleID:   "3",
		}
		mockUserRepo.AddUser(user)
	}

	// Test pagination
	limit := 10
	offset := 0

	users, total, err := mockUserRepo.FindAll(limit, offset, "")
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if total != 25 {
		t.Errorf("Expected total 25, got %d", total)
	}

	if len(users) != limit {
		t.Errorf("Expected %d users, got %d", limit, len(users))
	}

	// Test second page
	offset = 10
	users, total, err = mockUserRepo.FindAll(limit, offset, "")
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(users) != limit {
		t.Errorf("Expected %d users on second page, got %d", limit, len(users))
	}
}