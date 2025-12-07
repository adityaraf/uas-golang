package models

import "github.com/google/uuid"

type RolePermissions struct {
	RoleID       uuid.UUID `json:"role_id"`
	PermissionID uuid.UUID `json:"permission_id"`
}