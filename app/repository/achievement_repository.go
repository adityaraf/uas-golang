package repository

import (
	"context"
	models "crud-app/app/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository struct {
	collection *mongo.Collection
}

func NewAchievementRepository(db *mongo.Database) *AchievementRepository {
	return &AchievementRepository{
		collection: db.Collection("achievements"),
	}
}

// Create menyimpan achievement baru ke MongoDB
func (r *AchievementRepository) Create(ctx context.Context, achievement *models.Achievement) error {
	achievement.ID = primitive.NewObjectID()
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, achievement)
	return err
}

// FindByID mencari achievement berdasarkan achievement_id (exclude deleted)
func (r *AchievementRepository) FindByID(ctx context.Context, achievementID string) (*models.Achievement, error) {
	var achievement models.Achievement
	filter := bson.M{
		"achievement_id": achievementID,
		"is_deleted":     false,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&achievement)
	if err != nil {
		return nil, err
	}

	return &achievement, nil
}

// FindByStudentID mencari semua achievement berdasarkan student_id (exclude deleted)
func (r *AchievementRepository) FindByStudentID(ctx context.Context, studentID string) ([]models.Achievement, error) {
	var achievements []models.Achievement
	filter := bson.M{
		"student_id": studentID,
		"is_deleted": false,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}

// Update mengupdate achievement
func (r *AchievementRepository) Update(ctx context.Context, achievementID string, achievement *models.Achievement) error {
	achievement.UpdatedAt = time.Now()
	filter := bson.M{"achievement_id": achievementID}
	update := bson.M{"$set": achievement}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdateStatus mengupdate status achievement
func (r *AchievementRepository) UpdateStatus(ctx context.Context, achievementID string, status string) error {
	filter := bson.M{"achievement_id": achievementID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete menghapus achievement (hard delete - untuk rollback)
func (r *AchievementRepository) Delete(ctx context.Context, achievementID string) error {
	filter := bson.M{"achievement_id": achievementID}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// SoftDelete melakukan soft delete achievement (FR-005)
func (r *AchievementRepository) SoftDelete(ctx context.Context, achievementID string) error {
	filter := bson.M{
		"achievement_id": achievementID,
		"is_deleted":     false,
	}
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"updated_at": now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// FindAll mencari semua achievement dengan filter
func (r *AchievementRepository) FindAll(ctx context.Context, filter bson.M) ([]models.Achievement, error) {
	var achievements []models.Achievement

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}

// FindByAchievementIDs mencari achievements berdasarkan multiple achievement_ids (FR-006)
func (r *AchievementRepository) FindByAchievementIDs(ctx context.Context, achievementIDs []string) ([]models.Achievement, error) {
	var achievements []models.Achievement
	filter := bson.M{
		"achievement_id": bson.M{"$in": achievementIDs},
		"is_deleted":     false,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}