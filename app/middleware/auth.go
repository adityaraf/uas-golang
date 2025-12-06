package middleware

import (
	"crud-app/app/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Token akses diperlukan"})
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{"error": "Format token tidak valid"})
		}

		claims, err := utils.ValidateToken(tokenParts[1])
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Token tidak valid atau expired"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role_id", claims.RoleID)

		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID := c.Locals("role_id").(string)
		// Sesuaikan dengan ID role admin di database Anda
		if roleID != "1" { // Asumsi role_id admin = 1
			return c.Status(403).JSON(fiber.Map{"error": "Akses ditolak. Hanya admin yang diizinkan"})
		}
		return c.Next()
	}
}

func UserSelfOrAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		roleID := c.Locals("role_id").(string)

		// Get the ID from URL parameter
		paramID := c.Params("id")

		// Allow if user is admin or accessing their own resource
		if roleID == "1" || userID == paramID { // Asumsi role_id admin = 1
			return c.Next()
		}

		return c.Status(403).JSON(fiber.Map{"error": "Akses ditolak. Anda hanya dapat mengakses data sendiri"})
	}
}