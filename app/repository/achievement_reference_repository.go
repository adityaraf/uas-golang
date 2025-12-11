package repository

import (
models "crud-app/app/model"
"database/sql"
"fmt"
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

// FindByID mencari reference berdasarkan ID (exclude deleted)
func (r *AchievementReferenceRepository) FindByID(id string) (*models.AchievementReferences, error) {
query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       deleted_at, created_at, updated_at
		FROM achievement_references
		WHERE id = $1 AND deleted_at IS NULL
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
&ref.DeletedAt,
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

// FindByMongoID mencari reference berdasarkan mongo_achievement_id (exclude deleted)
func (r *AchievementReferenceRepository) FindByMongoID(mongoID string) (*models.AchievementReferences, error) {
query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       deleted_at, created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1 AND deleted_at IS NULL
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
&ref.DeletedAt,
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

// FindByStudentID mencari semua reference berdasarkan student_id (exclude deleted)
func (r *AchievementReferenceRepository) FindByStudentID(studentID string) ([]models.AchievementReferences, error) {
query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       deleted_at, created_at, updated_at
		FROM achievement_references
		WHERE student_id = $1 AND deleted_at IS NULL
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
&ref.DeletedAt,
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

// Delete menghapus reference (hard delete - untuk rollback)
func (r *AchievementReferenceRepository) Delete(mongoID string) error {
query := `DELETE FROM achievement_references WHERE mongo_achievement_id = $1`
_, err := r.db.Exec(query, mongoID)
return err
}

// SoftDelete melakukan soft delete reference (FR-005)
func (r *AchievementReferenceRepository) SoftDelete(mongoID string) error {
query := `
		UPDATE achievement_references
		SET deleted_at = $1, updated_at = $1
		WHERE mongo_achievement_id = $2 AND deleted_at IS NULL
	`

_, err := r.db.Exec(query, time.Now(), mongoID)
return err
}

// FindByStudentIDs mencari achievement references berdasarkan multiple student_ids (FR-006)
func (r *AchievementReferenceRepository) FindByStudentIDs(studentIDs []string, limit, offset int) ([]models.AchievementReferences, int64, error) {
if len(studentIDs) == 0 {
return []models.AchievementReferences{}, 0, nil
}

// Build placeholders for IN clause
placeholders := ""
args := make([]interface{}, 0)
for i, id := range studentIDs {
if i > 0 {
placeholders += ", "
}
placeholders += "$" + fmt.Sprintf("%d", i+1)
args = append(args, id)
}

// Count total
countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM achievement_references
		WHERE student_id::text IN (%s) AND deleted_at IS NULL
	`, placeholders)

var total int64
err := r.db.QueryRow(countQuery, args...).Scan(&total)
if err != nil {
return nil, 0, err
}

// Get data with pagination
query := fmt.Sprintf(`
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       deleted_at, created_at, updated_at
		FROM achievement_references
		WHERE student_id::text IN (%s) AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, placeholders, len(studentIDs)+1, len(studentIDs)+2)

args = append(args, limit, offset)
rows, err := r.db.Query(query, args...)
if err != nil {
return nil, 0, err
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
&ref.DeletedAt,
&ref.CreatedAt,
&ref.UpdatedAt,
)
if err != nil {
return nil, 0, err
}
references = append(references, ref)
}

return references, total, nil
}

// UpdateVerification mengupdate status menjadi verified dan set verified_by, verified_at (FR-007)
func (r *AchievementReferenceRepository) UpdateVerification(mongoID string, verifiedBy string, status string) error {
query := `
		UPDATE achievement_references
		SET status = $1, verified_by = $2, verified_at = $3, updated_at = $3
		WHERE mongo_achievement_id = $4
	`

verifiedByUUID, err := parseUUID(verifiedBy)
if err != nil {
return err
}

_, err = r.db.Exec(query, status, verifiedByUUID, time.Now(), mongoID)
return err
}

// UpdateRejection mengupdate status menjadi rejected dan set rejection_note (FR-007)
func (r *AchievementReferenceRepository) UpdateRejection(mongoID string, verifiedBy string, rejectionNote string) error {
query := `
		UPDATE achievement_references
		SET status = 'rejected', verified_by = $1, verified_at = $2, rejection_note = $3, updated_at = $2
		WHERE mongo_achievement_id = $4
	`

verifiedByUUID, err := parseUUID(verifiedBy)
if err != nil {
return err
}

_, err = r.db.Exec(query, verifiedByUUID, time.Now(), rejectionNote, mongoID)
return err
}

// FindPendingVerification mencari achievement yang perlu diverifikasi (status: submitted) (FR-007)
func (r *AchievementReferenceRepository) FindPendingVerification(limit, offset int) ([]models.AchievementReferences, int64, error) {
// Count total
countQuery := `
		SELECT COUNT(*)
		FROM achievement_references
		WHERE status = 'submitted' AND deleted_at IS NULL
	`

var total int64
err := r.db.QueryRow(countQuery).Scan(&total)
if err != nil {
return nil, 0, err
}

// Get data with pagination
query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       deleted_at, created_at, updated_at
		FROM achievement_references
		WHERE status = 'submitted' AND deleted_at IS NULL
		ORDER BY submitted_at ASC
		LIMIT $1 OFFSET $2
	`

rows, err := r.db.Query(query, limit, offset)
if err != nil {
return nil, 0, err
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
&ref.DeletedAt,
&ref.CreatedAt,
&ref.UpdatedAt,
)
if err != nil {
return nil, 0, err
}
references = append(references, ref)
}

return references, total, nil
}

// Helper function to parse UUID
func parseUUID(uuidStr string) (*string, error) {
if uuidStr == "" {
return nil, nil
}
return &uuidStr, nil
}

// FindAllWithFilters mencari semua achievement references dengan filter dan sorting (FR-010)
func (r *AchievementReferenceRepository) FindAllWithFilters(
	limit, offset int,
	statusFilter, studentIDFilter string,
	sortBy, sortOrder string,
) ([]models.AchievementReferences, int64, error) {
	// Build WHERE clause
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argIndex := 1

	if statusFilter != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, statusFilter)
		argIndex++
	}

	if studentIDFilter != "" {
		whereClause += fmt.Sprintf(" AND student_id::text = $%d", argIndex)
		args = append(args, studentIDFilter)
		argIndex++
	}

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM achievement_references
		%s
	`, whereClause)

	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderByClause := "ORDER BY created_at DESC" // default
	if sortBy != "" {
		validSortFields := map[string]bool{
			"created_at":   true,
			"submitted_at": true,
			"verified_at":  true,
			"updated_at":   true,
		}
		if validSortFields[sortBy] {
			order := "DESC"
			if sortOrder == "asc" {
				order = "ASC"
			}
			orderByClause = fmt.Sprintf("ORDER BY %s %s", sortBy, order)
		}
	}

	// Get data with pagination
	query := fmt.Sprintf(`
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       deleted_at, created_at, updated_at
		FROM achievement_references
		%s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, argIndex, argIndex+1)

	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
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
			&ref.DeletedAt,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		references = append(references, ref)
	}

	return references, total, nil
}