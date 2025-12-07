package repository

import (
	models "crud-app/app/model"
	"database/sql"
	"time"
)

type AchievementReferenceRepository struct {
	db *sql.DB
}

func NewAchievementReferenceRepository(db *sql.DB) *AchievementReferenceRepository {
	return &AchievementReferenceRepository{db: db}
}

// Create menyimpan reference achievement ke PostgreSQL
func (r *AchievementReferenceRepository) Create(ref *models.AchievementReferences) error {
	query := `
		INSERT INTO achievement_references 
		(id, student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		query,
		ref.ID,
		ref.StudentID,
		ref.MongoAchievementID,
		ref.Status,
		ref.CreatedAt,
		ref.UpdatedAt,
	)

	return err
}

// FindByID mencari reference berdasarkan ID
func (r *AchievementReferenceRepository) FindByID(id string) (*models.AchievementReferences, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		WHERE id = $1
	`

	var ref models.AchievementReferences
	err := r.db.QueryRow(query, id).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &ref, nil
}

// FindByMongoID mencari reference berdasarkan mongo_achievement_id
func (r *AchievementReferenceRepository) FindByMongoID(mongoID string) (*models.AchievementReferences, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1
	`

	var ref models.AchievementReferences
	err := r.db.QueryRow(query, mongoID).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &ref, nil
}

// FindByStudentID mencari semua reference berdasarkan student_id
func (r *AchievementReferenceRepository) FindByStudentID(studentID string) ([]models.AchievementReferences, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		WHERE student_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []models.AchievementReferences
	for rows.Next() {
		var ref models.AchievementReferences
		err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		references = append(references, ref)
	}

	return references, nil
}

// UpdateStatus mengupdate status reference
func (r *AchievementReferenceRepository) UpdateStatus(mongoID string, status string) error {
	query := `
		UPDATE achievement_references
		SET status = $1, updated_at = $2
		WHERE mongo_achievement_id = $3
	`

	_, err := r.db.Exec(query, status, time.Now(), mongoID)
	return err
}

// UpdateSubmittedStatus mengupdate status menjadi submitted dan set submitted_at
func (r *AchievementReferenceRepository) UpdateSubmittedStatus(mongoID string) error {
	query := `
		UPDATE achievement_references
		SET status = 'submitted', submitted_at = $1, updated_at = $1
		WHERE mongo_achievement_id = $2
	`

	_, err := r.db.Exec(query, time.Now(), mongoID)
	return err
}

// Delete menghapus reference
func (r *AchievementReferenceRepository) Delete(mongoID string) error {
	query := `DELETE FROM achievement_references WHERE mongo_achievement_id = $1`
	_, err := r.db.Exec(query, mongoID)
	return err
}