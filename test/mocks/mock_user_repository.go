package mocks

import (
	models "crud-app/app/model"
	"errors"
)

// MockUserRepository implements UserRepository interface for testing
type MockUserRepository struct {
	users map[string]*models.User
	calls map[string]int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
		calls: make(map[string]int),
	}
}

func (m *MockUserRepository) FindByUsernameOrEmail(identifier string) (*models.User, error) {
	m.calls["FindByUsernameOrEmail"]++

	for _, user := range m.users {
		if user.Username == identifier || user.Email == identifier {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetUserProfile(userID string) (*models.UserProfile, error) {
	m.calls["GetUserProfile"]++

	user, exists := m.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	return &models.UserProfile{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Email:    user.Email,
		RoleName: "Test Role",
	}, nil
}

func (m *MockUserRepository) Create(user *models.User) error {
	m.calls["Create"]++
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) FindByID(userID string) (*models.User, error) {
	m.calls["FindByID"]++

	user, exists := m.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) Update(userID string, user *models.User) error {
	m.calls["Update"]++

	if _, exists := m.users[userID]; !exists {
		return errors.New("user not found")
	}
	m.users[userID] = user
	return nil
}

func (m *MockUserRepository) SoftDelete(userID string) error {
	m.calls["SoftDelete"]++

	if _, exists := m.users[userID]; !exists {
		return errors.New("user not found")
	}
	delete(m.users, userID)
	return nil
}

func (m *MockUserRepository) CheckUsernameExists(username string) (bool, error) {
	m.calls["CheckUsernameExists"]++

	for _, user := range m.users {
		if user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) CheckEmailExists(email string) (bool, error) {
	m.calls["CheckEmailExists"]++

	for _, user := range m.users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) CheckRoleExists(roleID string) (bool, error) {
	m.calls["CheckRoleExists"]++
	// Mock: assume roles 1, 2, 3 exist
	return roleID == "1" || roleID == "2" || roleID == "3", nil
}

func (m *MockUserRepository) FindAll(limit, offset int, roleFilter string) ([]models.User, int64, error) {
	m.calls["FindAll"]++

	var users []models.User
	for _, user := range m.users {
		if roleFilter == "" || user.RoleID == roleFilter {
			users = append(users, *user)
		}
	}

	total := int64(len(users))

	// Apply pagination
	start := offset
	end := offset + limit
	if start > len(users) {
		return []models.User{}, total, nil
	}
	if end > len(users) {
		end = len(users)
	}

	return users[start:end], total, nil
}

func (m *MockUserRepository) AssignRole(userID string, roleID string) error {
	m.calls["AssignRole"]++

	user, exists := m.users[userID]
	if !exists {
		return errors.New("user not found")
	}
	user.RoleID = roleID
	return nil
}

func (m *MockUserRepository) UpdatePassword(userID string, passwordHash string) error {
	m.calls["UpdatePassword"]++

	user, exists := m.users[userID]
	if !exists {
		return errors.New("user not found")
	}
	user.PasswordHash = passwordHash
	return nil
}

func (m *MockUserRepository) FindAdvisorByStudentID(studentID string) (string, error) {
	m.calls["FindAdvisorByStudentID"]++
	// Mock implementation
	return "advisor-id", nil
}

// Helper methods for testing
func (m *MockUserRepository) AddUser(user *models.User) {
	m.users[user.ID] = user
}

func (m *MockUserRepository) GetCallCount(method string) int {
	return m.calls[method]
}

func (m *MockUserRepository) Reset() {
	m.users = make(map[string]*models.User)
	m.calls = make(map[string]int)
}