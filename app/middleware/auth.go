package middleware

import (
	"crud-app/app/utils"
	"strings"
	"strconv"

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
		c.Locals("role", claims.Role)

		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role").(string)
		if role != "admin" {
			return c.Status(403).JSON(fiber.Map{"error": "Akses ditolak. Hanya admin yang diizinkan"})
		}
		return c.Next()
	}
}

func UserSelfOrAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(int)
		role := c.Locals("role").(string)
		
		// Get the ID from URL parameter
		paramID, err := strconv.Atoi(c.Params("id"))
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
		}
		
		// Allow if user is admin or accessing their own resource
		if role == "admin" || userID == paramID {
			return c.Next()
		}
		
		return c.Status(403).JSON(fiber.Map{"error": "Akses ditolak. Anda hanya dapat mengakses data sendiri"})
	}
}
