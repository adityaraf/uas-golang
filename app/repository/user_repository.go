package repository

import (
	models "crud-app/app/model"
	"database/sql"
	"errors"
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