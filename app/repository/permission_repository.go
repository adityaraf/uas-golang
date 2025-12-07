package repository

import (
	"database/sql"
)

type PermissionRepository struct {
	db *sql.DB
}

func NewPermissionRepository(db *sql.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// GetUserPermissions mengambil semua permissions berdasarkan user ID
func (r *PermissionRepository) GetUserPermissions(userID string) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN users u ON u.role_id = rp.role_id
		WHERE u.id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// GetRolePermissions mengambil semua permissions berdasarkan role ID
func (r *PermissionRepository) GetRolePermissions(roleID string) ([]string, error) {
	query := `
		SELECT p.name
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`

	rows, err := r.db.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// CheckPermission mengecek apakah user memiliki permission tertentu
func (r *PermissionRepository) CheckPermission(userID string, permissionName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM permissions p
			INNER JOIN role_permissions rp ON p.id = rp.permission_id
			INNER JOIN users u ON u.role_id = rp.role_id
			WHERE u.id = $1 AND p.name = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(query, userID, permissionName).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}