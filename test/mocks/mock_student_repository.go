package mocks

import (
	models "crud-app/app/model"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockStudentRepository implements StudentRepository interface for testing
type MockStudentRepository struct {
	mock.Mock
	students map[string]*models.Student
}

func (m *MockStudentRepository) FindStudentIDsByAdvisorID(advisorID string) ([]string, error) {
	var studentIDs []string

	for _, student := range m.students {
		// Pastikan AdvisorID cocok.
		if student.AdvisorID == advisorID {
			// Mengembalikan UserID sebagai identifier relasi (sesuaikan jika logic Anda pakai ID lain)
			studentIDs = append(studentIDs, student.UserID)
		}
	}

	return studentIDs, nil
}

func NewMockStudentRepository() *MockStudentRepository {
	return &MockStudentRepository{
		students: make(map[string]*models.Student),
	}
}

// Create menyimpan data student baru ke memory map
func (m *MockStudentRepository) Create(student *models.Student) error {
	if student.ID == "" {
		student.ID = uuid.New().String()
	}
	m.students[student.ID] = student
	return nil
}

func (m *MockStudentRepository) AddStudent(student *models.Student) {
	if student.ID == "" {
		student.ID = uuid.New().String()
	}
	m.students[student.ID] = student
}

// GetStudentByID mencari student berdasarkan ID string
func (m *MockStudentRepository) GetStudentByID(id string) (*models.Student, error) {
	if student, exists := m.students[id]; exists {
		return student, nil
	}
	return nil, errors.New("student not found")
}

// GetStudentByUserID mencari student berdasarkan UserID
func (m *MockStudentRepository) GetStudentByUserID(userID string) (*models.Student, error) {
	for _, student := range m.students {
		if student.UserID == userID {
			return student, nil
		}
	}
	return nil, errors.New("student not found")
}

// GetStudentsByAdvisorID mencari list student berdasarkan AdvisorID
func (m *MockStudentRepository) GetStudentsByAdvisorID(advisorID string) ([]models.Student, error) {
	var results []models.Student

	for _, student := range m.students {
		// PERBAIKAN 6: AdvisorID bertipe string, cek kosong pakai "" bukan nil
		if student.AdvisorID != "" && student.AdvisorID == advisorID {
			results = append(results, *student)
		}
	}

	// Jika tidak ada error, return empty slice (bukan nil) agar aman
	return results, nil
}

// Update - Contoh method update jika diperlukan
func (m *MockStudentRepository) Update(student *models.Student) error {
	if _, exists := m.students[student.ID]; !exists {
		return errors.New("student not found")
	}

	// Update data di map
	m.students[student.ID] = student
	return nil
}

// Helper: Reset mock data (berguna untuk `defer` di unit test)
func (m *MockStudentRepository) Reset() {
	m.students = make(map[string]*models.Student)
}