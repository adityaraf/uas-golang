package repository

import (
models "crud-app/app/model"
"database/sql"
)

type LecturerRepository struct {
db *sql.DB
}

func NewLecturerRepository(db *sql.DB) *LecturerRepository {
return &LecturerRepository{db: db}
}

// Create membuat lecturer profile baru
func (r *LecturerRepository) Create(lecturer *models.Lecturer) error {
query := `
		INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

_, err := r.db.Exec(
query,
lecturer.ID,
lecturer.UserID,
lecturer.LecturerID,
lecturer.Department,
lecturer.CreatedAt,
)

return err
}

// FindByUserID mencari lecturer berdasarkan user_id
func (r *LecturerRepository) FindByUserID(userID string) (*models.Lecturer, error) {
query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE user_id = $1
	`

var lecturer models.Lecturer
err := r.db.QueryRow(query, userID).Scan(
&lecturer.ID,
&lecturer.UserID,
&lecturer.LecturerID,
&lecturer.Department,
&lecturer.CreatedAt,
)

if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}

return &lecturer, nil
}

// FindByID mencari lecturer berdasarkan ID
func (r *LecturerRepository) FindByID(id string) (*models.LecturerDetail, error) {
query := `
		SELECT l.id, l.user_id, l.lecturer_id, u.full_name, l.department
		FROM lecturers l
		INNER JOIN users u ON l.user_id = u.id
		WHERE l.id = $1
	`

var lecturer models.LecturerDetail
err := r.db.QueryRow(query, id).Scan(
&lecturer.ID,
&lecturer.UserID,
&lecturer.LecturerID,
&lecturer.FullName,
&lecturer.Department,
)

if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}

return &lecturer, nil
}

// Update mengupdate lecturer profile
func (r *LecturerRepository) Update(id string, lecturer *models.Lecturer) error {
query := `
		UPDATE lecturers
		SET lecturer_id = $1, department = $2
		WHERE id = $3
	`

_, err := r.db.Exec(
query,
lecturer.LecturerID,
lecturer.Department,
id,
)

return err
}

// Delete menghapus lecturer profile
func (r *LecturerRepository) Delete(id string) error {
query := `DELETE FROM lecturers WHERE id = $1`
_, err := r.db.Exec(query, id)
return err
}

// DeleteByUserID menghapus lecturer profile berdasarkan user_id
func (r *LecturerRepository) DeleteByUserID(userID string) error {
query := `DELETE FROM lecturers WHERE user_id = $1`
_, err := r.db.Exec(query, userID)
return err
}

// CheckExists mengecek apakah lecturer exists berdasarkan user_id
func (r *LecturerRepository) CheckExists(userID string) (bool, error) {
query := `SELECT EXISTS(SELECT 1 FROM lecturers WHERE user_id = $1)`

var exists bool
err := r.db.QueryRow(query, userID).Scan(&exists)
if err != nil {
return false, err
}

return exists, nil
}

// FindAll mencari semua lecturers dengan pagination
func (r *LecturerRepository) FindAll(limit, offset int) ([]models.LecturerDetail, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM lecturers`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get lecturers with pagination
	query := `
		SELECT l.id, l.user_id, l.lecturer_id, u.full_name, l.department
		FROM lecturers l
		INNER JOIN users u ON l.user_id = u.id
		ORDER BY u.full_name ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []models.LecturerDetail
	for rows.Next() {
		var lecturer models.LecturerDetail
		err := rows.Scan(
			&lecturer.ID,
			&lecturer.UserID,
			&lecturer.LecturerID,
			&lecturer.FullName,
			&lecturer.Department,
		)
		if err != nil {
			return nil, 0, err
		}
		lecturers = append(lecturers, lecturer)
	}

	return lecturers, total, nil
}