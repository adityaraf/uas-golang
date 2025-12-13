package mocks

import (
	"context"
	models "crud-app/app/model"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockAchievementRepository implements AchievementRepository interface for testing
type MockAchievementRepository struct {
	achievements map[string]*models.Achievement
	calls        map[string]int
}

func NewMockAchievementRepository() *MockAchievementRepository {
	return &MockAchievementRepository{
		achievements: make(map[string]*models.Achievement),
		calls:        make(map[string]int),
	}
}

func (m *MockAchievementRepository) Create(ctx context.Context, achievement *models.Achievement) error {
	m.calls["Create"]++

	if achievement.ID.IsZero() {
		achievement.ID = primitive.NewObjectID()
	}
	if achievement.CreatedAt.IsZero() {
		achievement.CreatedAt = time.Now()
	}
	achievement.UpdatedAt = time.Now()

	m.achievements[achievement.AchievementID] = achievement
	return nil
}

func (m *MockAchievementRepository) FindByID(ctx context.Context, achievementID string) (*models.Achievement, error) {
	m.calls["FindByID"]++

	achievement, exists := m.achievements[achievementID]
	if !exists || achievement.IsDeleted {
		return nil, errors.New("achievement not found")
	}
	return achievement, nil
}

func (m *MockAchievementRepository) FindByStudentID(ctx context.Context, studentID string) ([]models.Achievement, error) {
	m.calls["FindByStudentID"]++

	var results []models.Achievement
	for _, achievement := range m.achievements {
		if achievement.StudentID == studentID && !achievement.IsDeleted {
			results = append(results, *achievement)
		}
	}
	return results, nil
}

func (m *MockAchievementRepository) Update(ctx context.Context, achievementID string, achievement *models.Achievement) error {
	m.calls["Update"]++

	existing, exists := m.achievements[achievementID]
	if !exists || existing.IsDeleted {
		return errors.New("achievement not found")
	}

	achievement.UpdatedAt = time.Now()
	m.achievements[achievementID] = achievement
	return nil
}

func (m *MockAchievementRepository) UpdateStatus(ctx context.Context, achievementID string, status string) error {
	m.calls["UpdateStatus"]++

	achievement, exists := m.achievements[achievementID]
	if !exists || achievement.IsDeleted {
		return errors.New("achievement not found")
	}

	achievement.Status = status
	achievement.UpdatedAt = time.Now()
	return nil
}

func (m *MockAchievementRepository) Delete(ctx context.Context, achievementID string) error {
	m.calls["Delete"]++

	if _, exists := m.achievements[achievementID]; !exists {
		return errors.New("achievement not found")
	}

	delete(m.achievements, achievementID)
	return nil
}

func (m *MockAchievementRepository) SoftDelete(ctx context.Context, achievementID string) error {
	m.calls["SoftDelete"]++

	achievement, exists := m.achievements[achievementID]
	if !exists {
		return errors.New("achievement not found")
	}

	now := time.Now()
	achievement.IsDeleted = true
	achievement.DeletedAt = &now
	achievement.UpdatedAt = now
	return nil
}

func (m *MockAchievementRepository) FindAll(ctx context.Context, filter bson.M) ([]models.Achievement, error) {
	m.calls["FindAll"]++

	var results []models.Achievement
	for _, achievement := range m.achievements {
		// Simple filter implementation for testing
		if !achievement.IsDeleted {
			results = append(results, *achievement)
		}
	}
	return results, nil
}

func (m *MockAchievementRepository) FindByAchievementIDs(ctx context.Context, achievementIDs []string) ([]models.Achievement, error) {
	m.calls["FindByAchievementIDs"]++

	var results []models.Achievement
	for _, id := range achievementIDs {
		if achievement, exists := m.achievements[id]; exists && !achievement.IsDeleted {
			results = append(results, *achievement)
		}
	}
	return results, nil
}

func (m *MockAchievementRepository) GetStatisticsByStudentIDs(ctx context.Context, studentIDs []string) (map[string]interface{}, error) {
	m.calls["GetStatisticsByStudentIDs"]++

	stats := make(map[string]interface{})

	var filteredAchievements []models.Achievement
	for _, achievement := range m.achievements {
		if !achievement.IsDeleted {
			// Check if student ID is in the filter list
			for _, studentID := range studentIDs {
				if achievement.StudentID == studentID {
					filteredAchievements = append(filteredAchievements, *achievement)
					break
				}
			}
		}
	}

	// Calculate statistics
	totalAchievements := len(filteredAchievements)
	totalVerified := 0
	totalPending := 0
	totalRejected := 0
	totalDraft := 0

	categoryCount := make(map[string]int)
	levelCount := make(map[string]int)
	periodCount := make(map[string]int)

	for _, achievement := range filteredAchievements {
		switch achievement.Status {
		case "verified":
			totalVerified++
		case "submitted":
			totalPending++
		case "rejected":
			totalRejected++
		case "draft":
			totalDraft++
		}

		if achievement.Category != "" {
			categoryCount[achievement.Category]++
		}

		if achievement.Level != "" {
			levelCount[achievement.Level]++
		}

		yearMonth := achievement.Date.Format("2006-01")
		periodCount[yearMonth]++
	}

	stats["total_achievements"] = totalAchievements
	stats["total_verified"] = totalVerified
	stats["total_pending"] = totalPending
	stats["total_rejected"] = totalRejected
	stats["total_draft"] = totalDraft
	stats["category_count"] = categoryCount
	stats["level_count"] = levelCount
	stats["period_count"] = periodCount

	return stats, nil
}

// Helper methods for testing
func (m *MockAchievementRepository) AddAchievement(achievement *models.Achievement) {
	if achievement.ID.IsZero() {
		achievement.ID = primitive.NewObjectID()
	}
	m.achievements[achievement.AchievementID] = achievement
}

func (m *MockAchievementRepository) GetCallCount(method string) int {
	return m.calls[method]
}

func (m *MockAchievementRepository) Reset() {
	m.achievements = make(map[string]*models.Achievement)
	m.calls = make(map[string]int)
}

func (m *MockAchievementRepository) GetAchievementCount() int {
	count := 0
	for _, achievement := range m.achievements {
		if !achievement.IsDeleted {
			count++
		}
	}
	return count
}