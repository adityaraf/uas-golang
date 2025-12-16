package test

import (
	"context"
	models "crud-app/app/model"
	"crud-app/test/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAchievementService_Integration_SubmitForVerification(t *testing.T) {
	// Setup mocks
	mockAchievementRepo := mocks.NewMockAchievementRepository()
	mockReferenceRepo := mocks.NewMockAchievementReferenceRepository()

	// Generate Valid UUIDs
	achievementID := primitive.NewObjectID().Hex()
	userID := uuid.New().String()

	// Create test achievement in draft status
	achievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: achievementID,
		StudentID:     userID,
		Title:         "Test Achievement",
		Category:      "Kompetisi",
		Level:         "Nasional",
		Status:        "draft",
		Date:          time.Now(),
		IsDeleted:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	reference := &models.AchievementReferences{
		ID:                 uuid.New(),
		StudentID:          uuid.MustParse(userID),
		MongoAchievementID: achievementID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	mockAchievementRepo.AddAchievement(achievement)
	mockReferenceRepo.AddReference(reference)

	ctx := context.Background()

	// Test: Submit for verification
	err := mockAchievementRepo.UpdateStatus(ctx, achievementID, "submitted")
	if err != nil {
		t.Fatalf("UpdateStatus in MongoDB failed: %v", err)
	}

	err = mockReferenceRepo.UpdateSubmittedStatus(achievementID)
	if err != nil {
		t.Fatalf("UpdateSubmittedStatus in PostgreSQL failed: %v", err)
	}

	// Verify status changes
	updatedAchievement, err := mockAchievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if updatedAchievement.Status != "submitted" {
		t.Errorf("Expected status 'submitted', got '%s'", updatedAchievement.Status)
	}

	updatedReference, err := mockReferenceRepo.FindByMongoID(achievementID)
	if err != nil {
		t.Fatalf("FindByMongoID failed: %v", err)
	}

	if updatedReference.Status != "submitted" {
		t.Errorf("Expected reference status 'submitted', got '%s'", updatedReference.Status)
	}

	if updatedReference.SubmittedAt == nil {
		t.Error("Expected SubmittedAt to be set")
	}
}

func TestAchievementService_Integration_ApproveAchievement(t *testing.T) {
	// Setup mocks
	mockAchievementRepo := mocks.NewMockAchievementRepository()
	mockReferenceRepo := mocks.NewMockAchievementReferenceRepository()

	// Generate Valid UUIDs
	achievementID := primitive.NewObjectID().Hex()
	userID := uuid.New().String()
	lecturerID := uuid.New().String() // Use valid UUID string

	// Create test achievement in submitted status
	achievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: achievementID,
		StudentID:     userID,
		Title:         "Test Achievement",
		Status:        "submitted",
		Date:          time.Now(),
		IsDeleted:     false,
	}

	submittedAt := time.Now().Add(-1 * time.Hour)
	reference := &models.AchievementReferences{
		ID:                 uuid.New(),
		StudentID:          uuid.MustParse(userID),
		MongoAchievementID: achievementID,
		Status:             "submitted",
		SubmittedAt:        &submittedAt,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	mockAchievementRepo.AddAchievement(achievement)
	mockReferenceRepo.AddReference(reference)

	ctx := context.Background()

	// Test: Approve achievement
	err := mockAchievementRepo.UpdateStatus(ctx, achievementID, "verified")
	if err != nil {
		t.Fatalf("UpdateStatus in MongoDB failed: %v", err)
	}

	err = mockReferenceRepo.UpdateVerification(achievementID, lecturerID, "verified")
	if err != nil {
		t.Fatalf("UpdateVerification in PostgreSQL failed: %v", err)
	}

	// Verify approval
	updatedAchievement, err := mockAchievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if updatedAchievement.Status != "verified" {
		t.Errorf("Expected status 'verified', got '%s'", updatedAchievement.Status)
	}

	updatedReference, err := mockReferenceRepo.FindByMongoID(achievementID)
	if err != nil {
		t.Fatalf("FindByMongoID failed: %v", err)
	}

	if updatedReference.Status != "verified" {
		t.Errorf("Expected reference status 'verified', got '%s'", updatedReference.Status)
	}

	if updatedReference.VerifiedAt == nil {
		t.Error("Expected VerifiedAt to be set")
	}

	// FIX: Compare UUID string values
	if updatedReference.VerifiedBy == nil || updatedReference.VerifiedBy.String() != lecturerID {
		t.Errorf("Expected VerifiedBy to be '%s', got %v", lecturerID, updatedReference.VerifiedBy)
	}
}

func TestAchievementService_Integration_RejectAchievement(t *testing.T) {
	// Setup mocks
	mockAchievementRepo := mocks.NewMockAchievementRepository()
	mockReferenceRepo := mocks.NewMockAchievementReferenceRepository()

	// Generate Valid UUIDs
	achievementID := primitive.NewObjectID().Hex()
	userID := uuid.New().String()
	lecturerID := uuid.New().String()
	rejectionNote := "Dokumen tidak lengkap"

	achievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: achievementID,
		StudentID:     userID,
		Title:         "Test Achievement",
		Status:        "submitted",
		Date:          time.Now(),
		IsDeleted:     false,
	}

	reference := &models.AchievementReferences{
		ID:                 uuid.New(),
		StudentID:          uuid.MustParse(userID),
		MongoAchievementID: achievementID,
		Status:             "submitted",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	mockAchievementRepo.AddAchievement(achievement)
	mockReferenceRepo.AddReference(reference)

	ctx := context.Background()

	// Test: Reject achievement
	err := mockAchievementRepo.UpdateStatus(ctx, achievementID, "rejected")
	if err != nil {
		t.Fatalf("UpdateStatus in MongoDB failed: %v", err)
	}

	err = mockReferenceRepo.UpdateRejection(achievementID, lecturerID, rejectionNote)
	if err != nil {
		t.Fatalf("UpdateRejection in PostgreSQL failed: %v", err)
	}

	// Verify rejection
	updatedAchievement, err := mockAchievementRepo.FindByID(ctx, achievementID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if updatedAchievement.Status != "rejected" {
		t.Errorf("Expected status 'rejected', got '%s'", updatedAchievement.Status)
	}

	updatedReference, err := mockReferenceRepo.FindByMongoID(achievementID)
	if err != nil {
		t.Fatalf("FindByMongoID failed: %v", err)
	}

	if updatedReference.Status != "rejected" {
		t.Errorf("Expected reference status 'rejected', got '%s'", updatedReference.Status)
	}

	if updatedReference.RejectionNote == nil || *updatedReference.RejectionNote != rejectionNote {
		t.Errorf("Expected RejectionNote to be '%s', got %v", rejectionNote, updatedReference.RejectionNote)
	}
}

func TestAchievementService_Integration_GetAdviseeAchievements(t *testing.T) {
	// Setup mocks
	mockAchievementRepo := mocks.NewMockAchievementRepository()
	mockReferenceRepo := mocks.NewMockAchievementReferenceRepository()
	mockStudentRepo := mocks.NewMockStudentRepository()

	// Generate Valid UUID Strings
	advisorID := uuid.New().String()
	student1ID := uuid.New().String()
	student2ID := uuid.New().String()

	// Create students with advisor
	// FIX: Hapus field yang tidak dikenal (Year, StudentNumber) & ganti Major -> ProgramStudy
	student1 := &models.Student{
		ID:           uuid.New().String(),
		UserID:       student1ID,
		AdvisorID:    advisorID,
		ProgramStudy: "Computer Science", // Ganti dari Major
		CreatedAt:    time.Now(),
	}

	student2 := &models.Student{
		ID:           uuid.New().String(),
		UserID:       student2ID,
		AdvisorID:    advisorID,
		ProgramStudy: "Information Systems", // Ganti dari Major
		CreatedAt:    time.Now(),
	}

	// Method ini sekarang sudah tersedia di mock
	mockStudentRepo.AddStudent(student1)
	mockStudentRepo.AddStudent(student2)

	// Create achievements for students
	achievement1 := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: "achievement-1",
		StudentID:     student1ID,
		Title:         "Achievement 1",
		Status:        "verified",
		Date:          time.Now(),
		IsDeleted:     false,
	}

	achievement2 := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: "achievement-2",
		StudentID:     student2ID,
		Title:         "Achievement 2",
		Status:        "submitted",
		Date:          time.Now(),
		IsDeleted:     false,
	}

	mockAchievementRepo.AddAchievement(achievement1)
	mockAchievementRepo.AddAchievement(achievement2)

	// Create references
	reference1 := &models.AchievementReferences{
		ID:                 uuid.New(),
		StudentID:          uuid.MustParse(student1ID),
		MongoAchievementID: "achievement-1",
		Status:             "verified",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	reference2 := &models.AchievementReferences{
		ID:                 uuid.New(),
		StudentID:          uuid.MustParse(student2ID),
		MongoAchievementID: "achievement-2",
		Status:             "submitted",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	mockReferenceRepo.AddReference(reference1)
	mockReferenceRepo.AddReference(reference2)

	// Test: Get advisee achievements
	// Step 1: Get student IDs by advisor
	// Error 'cannot range over any' akan hilang karena return type mock sudah []string
	studentIDs, err := mockStudentRepo.FindStudentIDsByAdvisorID(advisorID)
	if err != nil {
		t.Fatalf("FindStudentIDsByAdvisorID failed: %v", err)
	}

	if len(studentIDs) != 2 {
		t.Errorf("Expected 2 students, got %d", len(studentIDs))
	}

	// Step 2: Get achievement references
	// Error 'cannot use as []string' akan hilang
	references, total, err := mockReferenceRepo.FindByStudentIDs(studentIDs, 10, 0)
	if err != nil {
		t.Fatalf("FindByStudentIDs failed: %v", err)
	}

	if len(references) != 2 {
		t.Errorf("Expected 2 references, got %d", len(references))
	}

	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}

	// Step 3: Get achievement details
	achievementIDs := make([]string, len(references))
	for i, ref := range references {
		achievementIDs[i] = ref.MongoAchievementID
	}

	ctx := context.Background()
	achievements, err := mockAchievementRepo.FindByAchievementIDs(ctx, achievementIDs)
	if err != nil {
		t.Fatalf("FindByAchievementIDs failed: %v", err)
	}

	if len(achievements) != 2 {
		t.Errorf("Expected 2 achievements, got %d", len(achievements))
	}

	for _, achievement := range achievements {
		found := false
		for _, studentID := range studentIDs {
			if achievement.StudentID == studentID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Achievement %s does not belong to any advisee", achievement.AchievementID)
		}
	}
}

func TestAchievementService_Integration_GetPendingVerification(t *testing.T) {
	mockAchievementRepo := mocks.NewMockAchievementRepository()
	mockReferenceRepo := mocks.NewMockAchievementReferenceRepository()

	// Generate Valid UUIDs
	student1 := uuid.New().String()
	student2 := uuid.New().String()
	student3 := uuid.New().String()

	achievements := []*models.Achievement{
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-1",
			StudentID:     student1,
			Title:         "Pending Achievement 1",
			Status:        "submitted",
			Date:          time.Now(),
			IsDeleted:     false,
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-2",
			StudentID:     student2,
			Title:         "Pending Achievement 2",
			Status:        "submitted",
			Date:          time.Now(),
			IsDeleted:     false,
		},
		{
			ID:            primitive.NewObjectID(),
			AchievementID: "achievement-3",
			StudentID:     student3,
			Title:         "Verified Achievement",
			Status:        "verified",
			Date:          time.Now(),
			IsDeleted:     false,
		},
	}

	references := []*models.AchievementReferences{
		{
			ID:                 uuid.New(),
			StudentID:          uuid.MustParse(student1),
			MongoAchievementID: "achievement-1",
			Status:             "submitted",
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			StudentID:          uuid.MustParse(student2),
			MongoAchievementID: "achievement-2",
			Status:             "submitted",
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		{
			ID:                 uuid.New(),
			StudentID:          uuid.MustParse(student3),
			MongoAchievementID: "achievement-3",
			Status:             "verified",
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}

	for _, achievement := range achievements {
		mockAchievementRepo.AddAchievement(achievement)
	}

	for _, reference := range references {
		mockReferenceRepo.AddReference(reference)
	}

	// Test: Get pending verification
	pendingReferences, total, err := mockReferenceRepo.FindPendingVerification(10, 0)
	if err != nil {
		t.Fatalf("FindPendingVerification failed: %v", err)
	}

	if len(pendingReferences) != 2 {
		t.Errorf("Expected 2 pending references, got %d", len(pendingReferences))
	}

	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}

	for _, ref := range pendingReferences {
		if ref.Status != "submitted" {
			t.Errorf("Expected status 'submitted', got '%s'", ref.Status)
		}
	}

	achievementIDs := make([]string, len(pendingReferences))
	for i, ref := range pendingReferences {
		achievementIDs[i] = ref.MongoAchievementID
	}

	ctx := context.Background()
	pendingAchievements, err := mockAchievementRepo.FindByAchievementIDs(ctx, achievementIDs)
	if err != nil {
		t.Fatalf("FindByAchievementIDs failed: %v", err)
	}

	if len(pendingAchievements) != 2 {
		t.Errorf("Expected 2 pending achievements, got %d", len(pendingAchievements))
	}

	for _, achievement := range pendingAchievements {
		if achievement.Status != "submitted" {
			t.Errorf("Expected achievement status 'submitted', got '%s'", achievement.Status)
		}
	}
}

func TestAchievementService_Integration_SoftDelete(t *testing.T) {
	mockAchievementRepo := mocks.NewMockAchievementRepository()
	mockReferenceRepo := mocks.NewMockAchievementReferenceRepository()

	achievementID := primitive.NewObjectID().Hex()
	userID := uuid.New().String()

	achievement := &models.Achievement{
		ID:            primitive.NewObjectID(),
		AchievementID: achievementID,
		StudentID:     userID,
		Title:         "Test Achievement",
		Status:        "draft",
		Date:          time.Now(),
		IsDeleted:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	reference := &models.AchievementReferences{
		ID:                 uuid.New(),
		StudentID:          uuid.MustParse(userID),
		MongoAchievementID: achievementID,
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	mockAchievementRepo.AddAchievement(achievement)
	mockReferenceRepo.AddReference(reference)

	ctx := context.Background()

	// Test: Soft delete
	err := mockAchievementRepo.SoftDelete(ctx, achievementID)
	if err != nil {
		t.Fatalf("SoftDelete in MongoDB failed: %v", err)
	}

	err = mockReferenceRepo.SoftDelete(achievementID)
	if err != nil {
		t.Fatalf("SoftDelete in PostgreSQL failed: %v", err)
	}

	// Verify soft deletion
	_, err = mockAchievementRepo.FindByID(ctx, achievementID)
	if err == nil {
		t.Error("Expected error when finding soft-deleted achievement")
	}

	_, err = mockReferenceRepo.FindByMongoID(achievementID)
	if err == nil {
		t.Error("Expected error when finding soft-deleted reference")
	}

	achievements, err := mockAchievementRepo.FindByStudentID(ctx, userID)
	if err != nil {
		t.Fatalf("FindByStudentID failed: %v", err)
	}

	if len(achievements) != 0 {
		t.Errorf("Expected 0 achievements after soft delete, got %d", len(achievements))
	}
}