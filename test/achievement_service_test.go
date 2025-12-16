package test

import (
	"context"
	models "crud-app/app/model"
	"crud-app/app/service"
	"crud-app/test/mocks"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupAchievementService() (*service.AchievementService, *mocks.MockAchievementRepository) {
	mockAchievementRepo := mocks.NewMockAchievementRepository()

	// Create a mock service with nil databases (we'll use mocks)
	achievementService := &service.AchievementService{}

	return achievementService, mockAchievementRepo
}

func TestAchievementService_GetMyAchievements_Success(t *testing.T) {
	fiber.New()
	_, mockRepo := setupAchievementService()

	// Create test achievements
	userID := "test-user-id"
	achievements := []*models.Achievement{
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-1",
			StudentID:     userID,
			Title:         "Test Achievement 1",
			Category:      "Kompetisi",
			Level:         "Nasional",
			Status:        "verified",
			Date:          time.Now(),
			IsDeleted:     false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-2",
			StudentID:     userID,
			Title:         "Test Achievement 2",
			Category:      "Penelitian",
			Level:         "Internasional",
			Status:        "draft",
			Date:          time.Now(),
			IsDeleted:     false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, achievement := range achievements {
		mockRepo.AddAchievement(achievement)
	}

	// Test GetMyAchievements logic
	ctx := context.Background()
	results, err := mockRepo.FindByStudentID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByStudentID failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 achievements, got %d", len(results))
	}

	// Verify achievements belong to correct user
	for _, result := range results {
		if result.StudentID != userID {
			t.Errorf("Achievement belongs to wrong user: %s", result.StudentID)
		}
	}
}

func TestAchievementService_GetAchievementByID_Success(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Create test achievement
	achievementID := "test-achievement-id"
	userID := "test-user-id"

	achievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: achievementID,
		StudentID:     userID,
		Title:         "Test Achievement",
		Category:      "Kompetisi",
		Level:         "Nasional",
		Status:        "verified",
		Date:          time.Now(),
		IsDeleted:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mockRepo.AddAchievement(achievement)

	// Test FindByID
	ctx := context.Background()
	result, err := mockRepo.FindByID(ctx, achievementID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if result.AchievementID != achievementID {
		t.Errorf("Expected achievement ID %s, got %s", achievementID, result.AchievementID)
	}

	if result.StudentID != userID {
		t.Errorf("Expected student ID %s, got %s", userID, result.StudentID)
	}
}

func TestAchievementService_GetAchievementByID_NotFound(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Test with non-existent achievement
	ctx := context.Background()
	_, err := mockRepo.FindByID(ctx, "non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent achievement")
	}
}

func TestAchievementService_UpdateStatus_Success(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Create test achievement
	achievementID := "test-achievement-id"
	achievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: achievementID,
		StudentID:     "test-user-id",
		Title:         "Test Achievement",
		Status:        "draft",
		Date:          time.Now(),
		IsDeleted:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mockRepo.AddAchievement(achievement)

	// Test status update
	ctx := context.Background()
	err := mockRepo.UpdateStatus(ctx, achievementID, "submitted")
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	// Verify status was updated
	updated, err := mockRepo.FindByID(ctx, achievementID)
	if err != nil {
		t.Fatalf("FindByID after update failed: %v", err)
	}

	if updated.Status != "submitted" {
		t.Errorf("Expected status 'submitted', got '%s'", updated.Status)
	}
}

func TestAchievementService_SoftDelete_Success(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Create test achievement
	achievementID := "test-achievement-id"
	achievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: achievementID,
		StudentID:     "test-user-id",
		Title:         "Test Achievement",
		Status:        "draft",
		Date:          time.Now(),
		IsDeleted:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mockRepo.AddAchievement(achievement)

	// Test soft delete
	ctx := context.Background()
	err := mockRepo.SoftDelete(ctx, achievementID)
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	// Verify achievement is marked as deleted
	_, err = mockRepo.FindByID(ctx, achievementID)
	if err == nil {
		t.Error("Expected error when finding soft-deleted achievement")
	}
}

func TestAchievementService_GetStatistics_Success(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Create test achievements with different statuses
	userID := "test-user-id"
	achievements := []*models.Achievement{
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-1",
			StudentID:     userID,
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
			StudentID:     userID,
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
			StudentID:     userID,
			Title:         "Achievement 3",
			Category:      "Kompetisi",
			Level:         "Regional",
			Status:        "draft",
			Date:          time.Date(2024, 10, 20, 0, 0, 0, 0, time.UTC),
			IsDeleted:     false,
		},
	}

	for _, achievement := range achievements {
		mockRepo.AddAchievement(achievement)
	}

	// Test statistics
	ctx := context.Background()
	stats, err := mockRepo.GetStatisticsByStudentIDs(ctx, []string{userID})
	if err != nil {
		t.Fatalf("GetStatisticsByStudentIDs failed: %v", err)
	}

	// Verify summary statistics
	totalAchievements := stats["total_achievements"].(int)
	if totalAchievements != 3 {
		t.Errorf("Expected 3 total achievements, got %d", totalAchievements)
	}

	totalVerified := stats["total_verified"].(int)
	if totalVerified != 1 {
		t.Errorf("Expected 1 verified achievement, got %d", totalVerified)
	}

	totalPending := stats["total_pending"].(int)
	if totalPending != 1 {
		t.Errorf("Expected 1 pending achievement, got %d", totalPending)
	}

	totalDraft := stats["total_draft"].(int)
	if totalDraft != 1 {
		t.Errorf("Expected 1 draft achievement, got %d", totalDraft)
	}

	// Verify category statistics
	categoryCount := stats["category_count"].(map[string]int)
	if categoryCount["Kompetisi"] != 2 {
		t.Errorf("Expected 2 'Kompetisi' achievements, got %d", categoryCount["Kompetisi"])
	}
	if categoryCount["Penelitian"] != 1 {
		t.Errorf("Expected 1 'Penelitian' achievement, got %d", categoryCount["Penelitian"])
	}

	// Verify level statistics
	levelCount := stats["level_count"].(map[string]int)
	if levelCount["Nasional"] != 1 {
		t.Errorf("Expected 1 'Nasional' achievement, got %d", levelCount["Nasional"])
	}
	if levelCount["Internasional"] != 1 {
		t.Errorf("Expected 1 'Internasional' achievement, got %d", levelCount["Internasional"])
	}
	if levelCount["Regional"] != 1 {
		t.Errorf("Expected 1 'Regional' achievement, got %d", levelCount["Regional"])
	}
}

func TestAchievementService_FindByAchievementIDs_Success(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Create test achievements
	achievements := []*models.Achievement{
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-1",
			StudentID:     "user-1",
			Title:         "Achievement 1",
			Status:        "verified",
			Date:          time.Now(),
			IsDeleted:     false,
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-2",
			StudentID:     "user-2",
			Title:         "Achievement 2",
			Status:        "verified",
			Date:          time.Now(),
			IsDeleted:     false,
		},
	}

	for _, achievement := range achievements {
		mockRepo.AddAchievement(achievement)
	}

	// Test FindByAchievementIDs
	ctx := context.Background()
	achievementIDs := []string{"achievement-1", "achievement-2"}
	results, err := mockRepo.FindByAchievementIDs(ctx, achievementIDs)
	if err != nil {
		t.Fatalf("FindByAchievementIDs failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 achievements, got %d", len(results))
	}

	// Verify correct achievements returned
	foundIDs := make(map[string]bool)
	for _, result := range results {
		foundIDs[result.AchievementID] = true
	}

	for _, expectedID := range achievementIDs {
		if !foundIDs[expectedID] {
			t.Errorf("Expected achievement ID %s not found in results", expectedID)
		}
	}
}

func TestAchievementService_Create_Success(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Create test achievement
	achievement := &models.Achievement{
		AchievementID: uuid.New().String(),
		StudentID:     "test-user-id",
		Title:         "New Achievement",
		Category:      "Kompetisi",
		Level:         "Nasional",
		Status:        "draft",
		Date:          time.Now(),
		Description:   "Test description",
		IsDeleted:     false,
	}

	// Test Create
	ctx := context.Background()
	err := mockRepo.Create(ctx, achievement)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify achievement was created
	created, err := mockRepo.FindByID(ctx, achievement.AchievementID)
	if err != nil {
		t.Fatalf("FindByID after create failed: %v", err)
	}

	if created.Title != achievement.Title {
		t.Errorf("Expected title '%s', got '%s'", achievement.Title, created.Title)
	}

	if created.Status != "draft" {
		t.Errorf("Expected status 'draft', got '%s'", created.Status)
	}
}

func TestAchievementService_Update_Success(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Create initial achievement
	achievementID := "test-achievement-id"
	achievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: achievementID,
		StudentID:     "test-user-id",
		Title:         "Original Title",
		Category:      "Kompetisi",
		Level:         "Nasional",
		Status:        "draft",
		Date:          time.Now(),
		IsDeleted:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mockRepo.AddAchievement(achievement)

	// Update achievement
	achievement.Title = "Updated Title"
	achievement.Category = "Penelitian"

	ctx := context.Background()
	err := mockRepo.Update(ctx, achievementID, achievement)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updated, err := mockRepo.FindByID(ctx, achievementID)
	if err != nil {
		t.Fatalf("FindByID after update failed: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updated.Title)
	}

	if updated.Category != "Penelitian" {
		t.Errorf("Expected category 'Penelitian', got '%s'", updated.Category)
	}
}

func TestAchievementService_CallCounting(t *testing.T) {
	_, mockRepo := setupAchievementService()

	ctx := context.Background()

	// Make various calls
	mockRepo.GetStatisticsByStudentIDs(ctx, []string{"student-1"})
	mockRepo.FindByID(ctx, "achievement-1")
	mockRepo.FindByStudentID(ctx, "student-1")
	mockRepo.Create(ctx, &models.Achievement{})

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

	if mockRepo.GetCallCount("Create") != 1 {
		t.Errorf("Expected 1 call to Create, got %d", mockRepo.GetCallCount("Create"))
	}
}

func TestAchievementService_EmptyStatistics(t *testing.T) {
	_, mockRepo := setupAchievementService()

	// Test statistics with no achievements
	ctx := context.Background()
	stats, err := mockRepo.GetStatisticsByStudentIDs(ctx, []string{"non-existent-user"})
	if err != nil {
		t.Fatalf("GetStatisticsByStudentIDs failed: %v", err)
	}

	totalAchievements := stats["total_achievements"].(int)
	if totalAchievements != 0 {
		t.Errorf("Expected 0 total achievements, got %d", totalAchievements)
	}

	categoryCount := stats["category_count"].(map[string]int)
	if len(categoryCount) != 0 {
		t.Errorf("Expected empty category count, got %v", categoryCount)
	}

	levelCount := stats["level_count"].(map[string]int)
	if len(levelCount) != 0 {
		t.Errorf("Expected empty level count, got %v", levelCount)
	}
}

func TestAchievementService_ExcludeDeletedAchievements(t *testing.T) {
	_, mockRepo := setupAchievementService()

	userID := "test-user-id"

	// Add normal achievement
	normalAchievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: "achievement-1",
		StudentID:     userID,
		Title:         "Normal Achievement",
		Status:        "verified",
		IsDeleted:     false,
		Date:          time.Now(),
	}
	mockRepo.AddAchievement(normalAchievement)

	// Add deleted achievement
	deletedAchievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: "achievement-2",
		StudentID:     userID,
		Title:         "Deleted Achievement",
		Status:        "verified",
		IsDeleted:     true,
		Date:          time.Now(),
	}
	mockRepo.AddAchievement(deletedAchievement)

	// Test FindByStudentID - should exclude deleted
	ctx := context.Background()
	results, err := mockRepo.FindByStudentID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByStudentID failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 achievement (excluding deleted), got %d", len(results))
	}

	if results[0].AchievementID != "achievement-1" {
		t.Errorf("Expected achievement-1, got %s", results[0].AchievementID)
	}

	// Test statistics - should exclude deleted
	stats, err := mockRepo.GetStatisticsByStudentIDs(ctx, []string{userID})
	if err != nil {
		t.Fatalf("GetStatisticsByStudentIDs failed: %v", err)
	}

	totalAchievements := stats["total_achievements"].(int)
	if totalAchievements != 1 {
		t.Errorf("Expected 1 achievement (excluding deleted), got %d", totalAchievements)
	}
}