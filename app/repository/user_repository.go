package repository

import (
models "crud-app/app/model"
"database/sql"
"errors"
	"time"
)

type UserRepository struct {
db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
return &UserRepository{db: db}
}

// FindByUsernameOrEmail mencari user berdasarkan username atau email
func (r *UserRepository) FindByUsernameOrEmail(identifier string) (*models.User, error) {
query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE username = $1 OR email = $1
		LIMIT 1
	`

var user models.User
err := r.db.QueryRow(query, identifier).Scan(
&user.ID,
&user.Username,
&user.Email,
&user.PasswordHash,
&user.FullName,
&user.RoleID,
&user.IsActive,
&user.CreatedAt,
&user.UpdatedAt,
)

if err == sql.ErrNoRows {
return nil, errors.New("user not found")
}
if err != nil {
return nil, err
}

return &user, nil
}

// GetUserProfile mengambil profile user dengan role name
func (r *UserRepository) GetUserProfile(userID string) (*models.UserProfile, error) {
query := `
		SELECT u.id, u.username, u.full_name, u.email, r.name as role_name
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`

var profile models.UserProfile
err := r.db.QueryRow(query, userID).Scan(
&profile.ID,
&profile.Username,
&profile.FullName,
&profile.Email,
&profile.RoleName,
)

if err != nil {
return nil, err
}

return &profile, nil
}

// FindAdvisorByStudentID mencari dosen wali berdasarkan student_id
func (r *UserRepository) FindAdvisorByStudentID(studentID string) (string, error) {
query := `
		SELECT advisor_id
		FROM students
		WHERE user_id = $1
		LIMIT 1
	`

var advisorID string
err := r.db.QueryRow(query, studentID).Scan(&advisorID)

if err == sql.ErrNoRows {
return "", errors.New("student not found or advisor not assigned")
}
if err != nil {
return "", err
}

return advisorID, nil
}

// Create membuat user baru (FR-009)
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.RoleID,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// FindAll mencari semua users dengan pagination (FR-009)
func (r *UserRepository) FindAll(limit, offset int, roleFilter string) ([]models.User, int64, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	args := []interface{}{}

	if roleFilter != "" {
		countQuery += ` AND role_id = $1`
		args = append(args, roleFilter)
	}

	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get data with pagination
	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
	`

	if roleFilter != "" {
		query += ` AND role_id = $1`
		query += ` ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.FullName,
			&user.RoleID,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	return users, total, nil
}

// FindByID mencari user berdasarkan ID (FR-009)
func (r *UserRepository) FindByID(userID string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.RoleID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update mengupdate user (FR-009)
func (r *UserRepository) Update(userID string, user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, full_name = $3, role_id = $4, is_active = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(
		query,
		user.Username,
		user.Email,
		user.FullName,
		user.RoleID,
		user.IsActive,
		user.UpdatedAt,
		userID,
	)

	return err
}

// UpdatePassword mengupdate password user
func (r *UserRepository) UpdatePassword(userID string, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, passwordHash, time.Now(), userID)
	return err
}

// AssignRole mengassign role ke user (FR-009)
func (r *UserRepository) AssignRole(userID string, roleID string) error {
	query := `
		UPDATE users
		SET role_id = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, roleID, time.Now(), userID)
	return err
}

// SoftDelete melakukan soft delete user (FR-009)
func (r *UserRepository) SoftDelete(userID string) error {
	query := `
		UPDATE users
		SET deleted_at = $1, updated_at = $1, is_active = false
		WHERE id = $2 AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}

// CheckUsernameExists mengecek apakah username sudah ada
func (r *UserRepository) CheckUsernameExists(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// CheckEmailExists mengecek apakah email sudah ada
func (r *UserRepository) CheckEmailExists(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// CheckRoleExists mengecek apakah role exists
func (r *UserRepository) CheckRoleExists(roleID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1)`

	var exists bool
	err := r.db.QueryRow(query, roleID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}