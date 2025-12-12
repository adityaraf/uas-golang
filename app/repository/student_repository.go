package repository

import (
models "crud-app/app/model"
"database/sql"
)

type StudentRepository struct {
db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
return &StudentRepository{db: db}
}

// Create membuat student profile baru
func (r *StudentRepository) Create(student *models.Student) error {
query := `
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

_, err := r.db.Exec(
query,
student.ID,
student.UserID,
student.StudentID,
student.ProgramStudy,
student.AcademicYear,
student.AdvisorID,
student.CreatedAt,
)

return err
}

// FindByUserID mencari student berdasarkan user_id
func (r *StudentRepository) FindByUserID(userID string) (*models.Student, error) {
query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE user_id = $1
	`

var student models.Student
err := r.db.QueryRow(query, userID).Scan(
&student.ID,
&student.UserID,
&student.StudentID,
&student.ProgramStudy,
&student.AcademicYear,
&student.AdvisorID,
&student.CreatedAt,
)

if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}

return &student, nil
}

// FindByID mencari student berdasarkan ID
func (r *StudentRepository) FindByID(id string) (*models.StudentDetail, error) {
query := `
		SELECT s.id, s.user_id, s.student_id, u.full_name, s.program_study, 
		       s.academic_year, s.advisor_id, 
		       COALESCE(u2.full_name, '') as advisor_name
		FROM students s
		INNER JOIN users u ON s.user_id = u.id
		LEFT JOIN users u2 ON s.advisor_id = u2.id
		WHERE s.id = $1
	`

var student models.StudentDetail
err := r.db.QueryRow(query, id).Scan(
&student.ID,
&student.UserID,
&student.StudentID,
&student.FullName,
&student.ProgramStudy,
&student.AcademicYear,
&student.AdvisorID,
&student.AdvisorName,
)

if err == sql.ErrNoRows {
return nil, nil
}
if err != nil {
return nil, err
}

return &student, nil
}

// Update mengupdate student profile
func (r *StudentRepository) Update(id string, student *models.Student) error {
query := `
		UPDATE students
		SET student_id = $1, program_study = $2, academic_year = $3, advisor_id = $4
		WHERE id = $5
	`

_, err := r.db.Exec(
query,
student.StudentID,
student.ProgramStudy,
student.AcademicYear,
student.AdvisorID,
id,
)

return err
}

// AssignAdvisor mengassign advisor ke student
func (r *StudentRepository) AssignAdvisor(studentID string, advisorID string) error {
query := `
		UPDATE students
		SET advisor_id = $1
		WHERE id = $2
	`

_, err := r.db.Exec(query, advisorID, studentID)
return err
}

// FindStudentIDsByAdvisorID mencari student IDs berdasarkan advisor_id
func (r *StudentRepository) FindStudentIDsByAdvisorID(advisorID string) ([]string, error) {
query := `
		SELECT user_id
		FROM students
		WHERE advisor_id = $1
	`

rows, err := r.db.Query(query, advisorID)
if err != nil {
return nil, err
}
defer rows.Close()

var studentIDs []string
for rows.Next() {
var userID string
if err := rows.Scan(&userID); err != nil {
return nil, err
}
studentIDs = append(studentIDs, userID)
}

return studentIDs, nil
}

// Delete menghapus student profile
func (r *StudentRepository) Delete(id string) error {
query := `DELETE FROM students WHERE id = $1`
_, err := r.db.Exec(query, id)
return err
}

// DeleteByUserID menghapus student profile berdasarkan user_id
func (r *StudentRepository) DeleteByUserID(userID string) error {
query := `DELETE FROM students WHERE user_id = $1`
_, err := r.db.Exec(query, userID)
return err
}

// FindAll mencari semua students dengan pagination
func (r *StudentRepository) FindAll(limit, offset int) ([]models.StudentDetail, int64, error) {
	// Get total count
	var total int64
	countQuery := `SELECT COUNT(*) FROM students`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get students with pagination
	query := `
		SELECT s.id, s.user_id, s.student_id, u.full_name, s.program_study, 
		       s.academic_year, s.advisor_id, 
		       COALESCE(u2.full_name, '') as advisor_name
		FROM students s
		INNER JOIN users u ON s.user_id = u.id
		LEFT JOIN users u2 ON s.advisor_id = u2.id
		ORDER BY s.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []models.StudentDetail
	for rows.Next() {
		var student models.StudentDetail
		err := rows.Scan(
			&student.ID,
			&student.UserID,
			&student.StudentID,
			&student.FullName,
			&student.ProgramStudy,
			&student.AcademicYear,
			&student.AdvisorID,
			&student.AdvisorName,
		)
		if err != nil {
			return nil, 0, err
		}
		students = append(students, student)
	}

	return students, total, nil
}

// FindByAdvisorID mencari students berdasarkan advisor_id
func (r *StudentRepository) FindByAdvisorID(advisorID string) ([]models.StudentDetail, error) {
	query := `
		SELECT s.id, s.user_id, s.student_id, u.full_name, s.program_study, 
		       s.academic_year, s.advisor_id, 
		       COALESCE(u2.full_name, '') as advisor_name
		FROM students s
		INNER JOIN users u ON s.user_id = u.id
		LEFT JOIN users u2 ON s.advisor_id = u2.id
		WHERE s.advisor_id = $1
		ORDER BY u.full_name ASC
	`

	rows, err := r.db.Query(query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.StudentDetail
	for rows.Next() {
		var student models.StudentDetail
		err := rows.Scan(
			&student.ID,
			&student.UserID,
			&student.StudentID,
			&student.FullName,
			&student.ProgramStudy,
			&student.AcademicYear,
			&student.AdvisorID,
			&student.AdvisorName,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}