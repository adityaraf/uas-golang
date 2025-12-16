package mocks

import (
    models "crud-app/app/model"
    "errors"
    "time"

    "github.com/google/uuid"
)

// MockAchievementReferenceRepository implements AchievementReferenceRepository interface for testing
type MockAchievementReferenceRepository struct {
    references map[string]*models.AchievementReferences
    calls      map[string]int
}

func NewMockAchievementReferenceRepository() *MockAchievementReferenceRepository {
    return &MockAchievementReferenceRepository{
        references: make(map[string]*models.AchievementReferences),
        calls:      make(map[string]int),
    }
}

func (m *MockAchievementReferenceRepository) Create(reference *models.AchievementReferences) error {
    m.calls["Create"]++

    if reference.ID == uuid.Nil {
        reference.ID = uuid.New()
    }
    if reference.CreatedAt.IsZero() {
        reference.CreatedAt = time.Now()
    }
    reference.UpdatedAt = time.Now()

    m.references[reference.MongoAchievementID] = reference
    return nil
}

func (m *MockAchievementReferenceRepository) FindByMongoID(mongoID string) (*models.AchievementReferences, error) {
    m.calls["FindByMongoID"]++

    reference, exists := m.references[mongoID]
    // Change: Check if DeletedAt is not nil (meaning it is deleted)
    if !exists || reference.DeletedAt != nil {
        return nil, errors.New("reference not found")
    }
    return reference, nil
}

func (m *MockAchievementReferenceRepository) FindByStudentIDs(studentIDs []string, limit, offset int) ([]models.AchievementReferences, int64, error) {
    m.calls["FindByStudentIDs"]++

    var results []models.AchievementReferences
    count := 0

    for _, ref := range m.references {
        // Change: Skip if DeletedAt is not nil
        if ref.DeletedAt != nil {
            continue
        }

        // Check if student ID matches
        for _, studentID := range studentIDs {
            if ref.StudentID.String() == studentID {
                if count >= offset && len(results) < limit {
                    results = append(results, *ref)
                }
                count++
                break
            }
        }
    }

    return results, int64(count), nil
}

func (m *MockAchievementReferenceRepository) FindPendingVerification(limit, offset int) ([]models.AchievementReferences, int64, error) {
    m.calls["FindPendingVerification"]++

    var results []models.AchievementReferences
    count := 0

    for _, ref := range m.references {
        // Change: Check ref.DeletedAt == nil
        if ref.Status == "submitted" && ref.DeletedAt == nil {
            if count >= offset && len(results) < limit {
                results = append(results, *ref)
            }
            count++
        }
    }

    return results, int64(count), nil
}

func (m *MockAchievementReferenceRepository) FindAllWithFilters(limit, offset int, statusFilter, studentIDFilter, sortBy, sortOrder string) ([]models.AchievementReferences, int64, error) {
    m.calls["FindAllWithFilters"]++

    var results []models.AchievementReferences
    count := 0

    for _, ref := range m.references {
        // Change: Skip if deleted
        if ref.DeletedAt != nil {
            continue
        }

        // Apply filters
        if statusFilter != "" && ref.Status != statusFilter {
            continue
        }
        if studentIDFilter != "" && ref.StudentID.String() != studentIDFilter {
            continue
        }

        if count >= offset && len(results) < limit {
            results = append(results, *ref)
        }
        count++
    }

    return results, int64(count), nil
}

func (m *MockAchievementReferenceRepository) UpdateSubmittedStatus(mongoID string) error {
    m.calls["UpdateSubmittedStatus"]++

    reference, exists := m.references[mongoID]
    // Change: Check DeletedAt
    if !exists || reference.DeletedAt != nil {
        return errors.New("reference not found")
    }

    now := time.Now()
    reference.Status = "submitted"
    reference.SubmittedAt = &now
    reference.UpdatedAt = now
    return nil
}

func (m *MockAchievementReferenceRepository) UpdateVerification(mongoID, verifierID, status string) error {
    m.calls["UpdateVerification"]++

    reference, exists := m.references[mongoID]
    if !exists || reference.DeletedAt != nil {
        return errors.New("reference not found")
    }

    // PERBAIKAN: Parse string ke UUID
    verifierUUID, err := uuid.Parse(verifierID)
    if err != nil {
        return errors.New("invalid verifier UUID format")
    }

    now := time.Now()
    reference.Status = status
    reference.VerifiedAt = &now
    reference.VerifiedBy = &verifierUUID // Sekarang tipenya sudah benar (*uuid.UUID)
    reference.UpdatedAt = now
    return nil
}

func (m *MockAchievementReferenceRepository) UpdateRejection(mongoID, rejectorID, rejectionNote string) error {
    m.calls["UpdateRejection"]++

    reference, exists := m.references[mongoID]
    if !exists || reference.DeletedAt != nil {
        return errors.New("reference not found")
    }

    // PERBAIKAN: Parse string ke UUID (untuk disimpan di VerifiedBy atau field sejenis)
    // Sesuai logic SQL repository Anda sebelumnya, rejector juga disimpan di VerifiedBy
    rejectorUUID, err := uuid.Parse(rejectorID)
    if err != nil {
        return errors.New("invalid rejector UUID format")
    }

    now := time.Now()
    reference.Status = "rejected"
    reference.RejectionNote = &rejectionNote
    reference.VerifiedBy = &rejectorUUID // Update field verifier/rejector
    reference.UpdatedAt = now
    return nil
}

func (m *MockAchievementReferenceRepository) SoftDelete(mongoID string) error {
    m.calls["SoftDelete"]++

    reference, exists := m.references[mongoID]
    if !exists {
        return errors.New("reference not found")
    }

    now := time.Now()
    // Change: Remove IsDeleted = true, just set DeletedAt
    reference.DeletedAt = &now
    reference.UpdatedAt = now
    return nil
}

func (m *MockAchievementReferenceRepository) GetTopStudents(studentIDs []string, limit int) ([]models.TopStudent, error) {
    m.calls["GetTopStudents"]++

    studentStats := make(map[string]*models.TopStudent)

    for _, ref := range m.references {
        // Change: Skip if deleted
        if ref.DeletedAt != nil {
            continue
        }

        studentIDStr := ref.StudentID.String()

        found := false
        for _, id := range studentIDs {
            if id == studentIDStr {
                found = true
                break
            }
        }
        if !found {
            continue
        }

        if _, exists := studentStats[studentIDStr]; !exists {
            studentStats[studentIDStr] = &models.TopStudent{
                StudentID:   studentIDStr,
                StudentName: "Test Student " + studentIDStr,
            }
        }

        studentStats[studentIDStr].TotalAchievements++
        if ref.Status == "verified" {
            studentStats[studentIDStr].VerifiedAchievements++
        }
    }

    var results []models.TopStudent
    for _, student := range studentStats {
        results = append(results, *student)
        if len(results) >= limit {
            break
        }
    }

    return results, nil
}

func (m *MockAchievementReferenceRepository) GetAllTopStudents(limit int) ([]models.TopStudent, error) {
    m.calls["GetAllTopStudents"]++

    studentStats := make(map[string]*models.TopStudent)

    for _, ref := range m.references {
        // Change: Skip if deleted
        if ref.DeletedAt != nil {
            continue
        }

        studentIDStr := ref.StudentID.String()
        if _, exists := studentStats[studentIDStr]; !exists {
            studentStats[studentIDStr] = &models.TopStudent{
                StudentID:   studentIDStr,
                StudentName: "Test Student " + studentIDStr,
            }
        }

        studentStats[studentIDStr].TotalAchievements++
        if ref.Status == "verified" {
            studentStats[studentIDStr].VerifiedAchievements++
        }
    }

    var results []models.TopStudent
    for _, student := range studentStats {
        results = append(results, *student)
        if len(results) >= limit {
            break
        }
    }

    return results, nil
}

// Helper methods for testing
func (m *MockAchievementReferenceRepository) AddReference(reference *models.AchievementReferences) {
    if reference.ID == uuid.Nil {
        reference.ID = uuid.New()
    }
    m.references[reference.MongoAchievementID] = reference
}

func (m *MockAchievementReferenceRepository) GetCallCount(method string) int {
    return m.calls[method]
}

func (m *MockAchievementReferenceRepository) Reset() {
    m.references = make(map[string]*models.AchievementReferences)
    m.calls = make(map[string]int)
}

func (m *MockAchievementReferenceRepository) GetReferenceCount() int {
    count := 0
    for _, ref := range m.references {
        // Change: Check DeletedAt
        if ref.DeletedAt == nil {
            count++
        }
    }
    return count
}