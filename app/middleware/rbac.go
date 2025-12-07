package middleware

import (
	"crud-app/app/repository"
	"crud-app/app/utils"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RBACMiddleware struct {
	permRepo *repository.PermissionRepository
}

func NewRBACMiddleware(db *sql.DB) *RBACMiddleware {
	return &RBACMiddleware{
		permRepo: repository.NewPermissionRepository(db),
	}
}

// RequirePermission middleware untuk mengecek permission
func (m *RBACMiddleware) RequirePermission(permissionName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Step 1: Ekstrak user_id dari context (sudah di-set oleh AuthRequired middleware)
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Unauthorized: User ID tidak ditemukan",
			})
		}

		// Step 2: Cek cache terlebih dahulu
		cacheKey := fmt.Sprintf("user_permissions:%s", userID)
		var permissions []string

		if cachedPerms, found := utils.Cache.Get(cacheKey); found {
			permissions = cachedPerms.([]string)
		} else {
			// Step 3: Load permissions dari database jika tidak ada di cache
			perms, err := m.permRepo.GetUserPermissions(userID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"status":  "error",
					"message": "Gagal mengambil permissions",
				})
			}
			permissions = perms

			// Step 4: Simpan ke cache dengan TTL 15 menit
			utils.Cache.Set(cacheKey, permissions, 15*time.Minute)
		}

		// Step 5: Check apakah user memiliki permission yang diperlukan
		hasPermission := false
		for _, perm := range permissions {
			if perm == permissionName {
				hasPermission = true
				break
			}
		}

		// Step 6: Allow/deny request
		if !hasPermission {
			return c.Status(403).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Forbidden: Anda tidak memiliki permission '%s'", permissionName),
			})
		}

		return c.Next()
	}
}

// RequireAnyPermission middleware untuk mengecek salah satu dari beberapa permissions
func (m *RBACMiddleware) RequireAnyPermission(permissionNames ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Unauthorized: User ID tidak ditemukan",
			})
		}

		cacheKey := fmt.Sprintf("user_permissions:%s", userID)
		var permissions []string

		if cachedPerms, found := utils.Cache.Get(cacheKey); found {
			permissions = cachedPerms.([]string)
		} else {
			perms, err := m.permRepo.GetUserPermissions(userID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"status":  "error",
					"message": "Gagal mengambil permissions",
				})
			}
			permissions = perms
			utils.Cache.Set(cacheKey, permissions, 15*time.Minute)
		}

		// Check apakah user memiliki salah satu permission
		hasPermission := false
		for _, perm := range permissions {
			for _, requiredPerm := range permissionNames {
				if perm == requiredPerm {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if !hasPermission {
			return c.Status(403).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Forbidden: Anda tidak memiliki salah satu dari permissions: %v", permissionNames),
			})
		}

		return c.Next()
	}
}

// RequireAllPermissions middleware untuk mengecek semua permissions harus dimiliki
func (m *RBACMiddleware) RequireAllPermissions(permissionNames ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return c.Status(401).JSON(fiber.Map{
				"status":  "error",
				"message": "Unauthorized: User ID tidak ditemukan",
			})
		}

		cacheKey := fmt.Sprintf("user_permissions:%s", userID)
		var permissions []string

		if cachedPerms, found := utils.Cache.Get(cacheKey); found {
			permissions = cachedPerms.([]string)
		} else {
			perms, err := m.permRepo.GetUserPermissions(userID)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"status":  "error",
					"message": "Gagal mengambil permissions",
				})
			}
			permissions = perms
			utils.Cache.Set(cacheKey, permissions, 15*time.Minute)
		}

		// Check apakah user memiliki semua permissions yang diperlukan
		permMap := make(map[string]bool)
		for _, perm := range permissions {
			permMap[perm] = true
		}

		missingPerms := []string{}
		for _, requiredPerm := range permissionNames {
			if !permMap[requiredPerm] {
				missingPerms = append(missingPerms, requiredPerm)
			}
		}

		if len(missingPerms) > 0 {
			return c.Status(403).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Forbidden: Anda tidak memiliki permissions: %v", missingPerms),
			})
		}

		return c.Next()
	}
}