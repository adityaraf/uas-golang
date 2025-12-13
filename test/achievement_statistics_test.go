package test

import (
	"context"
	models "crud-app/app/model"
	"crud-app/test/mocks"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAchievementStatistics_GetStatisticsByStudentIDs(t *testing.T) {
	mockRepo := mocks.NewMockAchievementRepository()

	// Create test achievements
	studentID1 := "student-1"
	studentID2 := "student-2"

	achievements := []*models.Achievement{
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-1",
			StudentID:     studentID1,
			Title:         "Achievement 1",
			Category:      "Kompetisi",
			Level:         "Nasional",
			Status:        "verified",
			Date:          time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			IsDeleted:     false,
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-2",
			StudentID:     studentID1,
			Title:         "Achievement 2",
			Category:      "Penelitian",
			Level:         "Internasional",
			Status:        "submitted",
			Date:          time.Date(2024, 11, 15, 0, 0, 0, 0, time.UTC),
			IsDeleted:     false,
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-3",
			StudentID:     studentID2,
			Title:         "Achievement 3",
			Category:      "Kompetisi",
			Level:         "Regional",
			Status:        "rejected",
			Date:          time.Date(2024, 10, 20, 0, 0, 0, 0, time.UTC),
			IsDeleted:     false,
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-4",
			StudentID:     studentID1,
			Title:         "Achievement 4",
			Category:      "Kompetisi",
			Level:         "Nasional",
			Status:        "draft",
			Date:          time.Date(2024, 12, 5, 0, 0, 0, 0, time.UTC),
			IsDeleted:     false,
		},
	}

	// Add achievements to mock repository
	for _, achievement := range achievements {
		mockRepo.AddAchievement(achievement)
	}

	// Test statistics for student 1
	ctx := context.Background()
	stats, err := mockRepo.GetStatisticsByStudentIDs(ctx, []string{studentID1})
	if err != nil {
		t.Fatalf("GetStatisticsByStudentIDs failed: %v", err)
	}

	// Verify summary statistics
	totalAchievements := stats["total_achievements"].(int)
	if totalAchievements != 3 {
		t.Errorf("Expected 3 total achievements for student 1, got %d", totalAchievements)
	}

	totalVerified := stats["total_verified"].(int)
	if totalVerified != 1 {
		t.Errorf("Expected 1 verified achievement for student 1, got %d", totalVerified)
	}

	totalPending := stats["total_pending"].(int)
	if totalPending != 1 {
		t.Errorf("Expected 1 pending achievement for student 1, got %d", totalPending)
	}

	totalDraft := stats["total_draft"].(int)
	if totalDraft != 1 {
		t.Errorf("Expected 1 draft achievement for student 1, got %d", totalDraft)
	}

	// Verify category statistics
	categoryCount := stats["category_count"].(map[string]int)
	if categoryCount["Kompetisi"] != 2 {
		t.Errorf("Expected 2 'Kompetisi' achievements for student 1, got %d", categoryCount["Kompetisi"])
	}
	if categoryCount["Penelitian"] != 1 {
		t.Errorf("Expected 1 'Penelitian' achievement for student 1, got %d", categoryCount["Penelitian"])
	}

	// Verify level statistics
	levelCount := stats["level_count"].(map[string]int)
	if levelCount["Nasional"] != 2 {
		t.Errorf("Expected 2 'Nasional' achievements for student 1, got %d", levelCount["Nasional"])
	}
	if levelCount["Internasional"] != 1 {
		t.Errorf("Expected 1 'Internasional' achievement for student 1, got %d", levelCount["Internasional"])
	}
}

func TestAchievementStatistics_MultipleStudents(t *testing.T) {
	mockRepo := mocks.NewMockAchievementRepository()

	studentID1 := "student-1"
	studentID2 := "student-2"

	// Add achievements for both students
	achievements := []*models.Achievement{
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-1",
			StudentID:     studentID1,
			Category:      "Kompetisi",
			Level:         "Nasional",
			Status:        "verified",
			Date:          time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-2",
			StudentID:     studentID2,
			Category:      "Penelitian",
			Level:         "Internasional",
			Status:        "verified",
			Date:          time.Date(2024, 11, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, achievement := range achievements {
		mockRepo.AddAchievement(achievement)
	}

	// Test statistics for both students
	ctx := context.Background()
	stats, err := mockRepo.GetStatisticsByStudentIDs(ctx, []string{studentID1, studentID2})
	if err != nil {
		t.Fatalf("GetStatisticsByStudentIDs failed: %v", err)
	}

	totalAchievements := stats["total_achievements"].(int)
	if totalAchievements != 2 {
		t.Errorf("Expected 2 total achievements for both students, got %d", totalAchievements)
	}

	totalVerified := stats["total_verified"].(int)
	if totalVerified != 2 {
		t.Errorf("Expected 2 verified achievements for both students, got %d", totalVerified)
	}
}

func TestAchievementStatistics_EmptyResult(t *testing.T) {
	mockRepo := mocks.NewMockAchievementRepository()

	// Test with non-existent student
	ctx := context.Background()
	stats, err := mockRepo.GetStatisticsByStudentIDs(ctx, []string{"non-existent-student"})
	if err != nil {
		t.Fatalf("GetStatisticsByStudentIDs failed: %v", err)
	}

	totalAchievements := stats["total_achievements"].(int)
	if totalAchievements != 0 {
		t.Errorf("Expected 0 total achievements for non-existent student, got %d", totalAchievements)
	}

	categoryCount := stats["category_count"].(map[string]int)
	if len(categoryCount) != 0 {
		t.Errorf("Expected empty category count, got %v", categoryCount)
	}
}

func TestAchievementStatistics_ExcludeDeleted(t *testing.T) {
	mockRepo := mocks.NewMockAchievementRepository()

	studentID := "student-1"

	// Add normal achievement
	normalAchievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: "achievement-1",
		StudentID:     studentID,
		Status:        "verified",
		IsDeleted:     false,
		Date:          time.Now(),
	}
	mockRepo.AddAchievement(normalAchievement)

	// Add deleted achievement
	deletedAchievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: "achievement-2",
		StudentID:     studentID,
		Status:        "verified",
		IsDeleted:     true,
		Date:          time.Now(),
	}
	mockRepo.AddAchievement(deletedAchievement)

	// Test statistics - should exclude deleted
	ctx := context.Background()
	stats, err := mockRepo.GetStatisticsByStudentIDs(ctx, []string{studentID})
	if err != nil {
		t.Fatalf("GetStatisticsByStudentIDs failed: %v", err)
	}

	totalAchievements := stats["total_achievements"].(int)
	if totalAchievements != 1 {
		t.Errorf("Expected 1 achievement (excluding deleted), got %d", totalAchievements)
	}
}

func TestAchievementStatistics_PeriodGrouping(t *testing.T) {
	mockRepo := mocks.NewMockAchievementRepository()

	studentID := "student-1"

	// Add achievements in different months
	achievements := []*models.Achievement{
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-1",
			StudentID:     studentID,
			Status:        "verified",
			Date:          time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-2",
			StudentID:     studentID,
			Status:        "verified",
			Date:          time.Date(2024, 12, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-3",
			StudentID:     studentID,
			Status:        "verified",
			Date:          time.Date(2024, 11, 10, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, achievement := range achievements {
		mockRepo.AddAchievement(achievement)
	}

	// Test period statistics
	ctx := context.Background()
	stats, err := mockRepo.GetStatisticsByStudentIDs(ctx, []string{studentID})
	if err != nil {
		t.Fatalf("GetStatisticsByStudentIDs failed: %v", err)
	}

	periodCount := stats["period_count"].(map[string]int)

	// Should have 2 achievements in 2024-12 and 1 in 2024-11
	if periodCount["2024-12"] != 2 {
		t.Errorf("Expected 2 achievements in 2024-12, got %d", periodCount["2024-12"])
	}
	if periodCount["2024-11"] != 1 {
		t.Errorf("Expected 1 achievement in 2024-11, got %d", periodCount["2024-11"])
	}
}

func TestAchievementRepository_CallCounting(t *testing.T) {
	mockRepo := mocks.NewMockAchievementRepository()

	// Test that calls are being counted
	ctx := context.Background()

	// Make some calls
	mockRepo.GetStatisticsByStudentIDs(ctx, []string{"student-1"})
	mockRepo.FindByID(ctx, "achievement-1")
	mockRepo.FindByStudentID(ctx, "student-1")

	// Verify call counts
	if mockRepo.GetCallCount("GetStatisticsByStudentIDs") != 1 {
		t.Errorf("Expected 1 call to GetStatisticsByStudentIDs, got %d", mockRepo.GetCallCount("GetStatisticsByStudentIDs"))
	}
	if mockRepo.GetCallCount("FindByID") != 1 {
		t.Errorf("Expected 1 call to FindByID, got %d", mockRepo.GetCallCount("FindByID"))
	}
	if mockRepo.GetCallCount("FindByStudentID") != 1 {
		t.Errorf("Expected 1 call to FindByStudentID, got %d", mockRepo.GetCallCount("FindByStudentID"))
	}
}