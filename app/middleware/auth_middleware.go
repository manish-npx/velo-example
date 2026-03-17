// app/middleware/auth_middleware.go
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/manish-npx/velo/core"
)

// AuthMiddleware validates a JWT Bearer token.
// Returns a fiber.Handler — compatible with router.Use() and route-level middleware.
//
// BUG FIX: The original imported "github.com/golang-jwt/jwt/v5" correctly but
// used jwt.MapClaims as map[string]interface{} in AdminMiddleware — those are
// distinct types and the type assertion would panic at runtime. Fixed below.
func AuthMiddleware(app *core.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Missing authorization header"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid authorization header format"})
		}

		secret := app.Get("jwt_secret")
		if secret == nil {
			return c.Status(500).JSON(fiber.Map{"error": "JWT secret not configured"})
		}

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(401, "unexpected signing method")
			}
			return []byte(secret.(string)), nil
		})
		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired token"})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("user", claims)
			if id, ok := claims["id"].(float64); ok {
				c.Locals("user_id", uint(id))
			}
		}
		return c.Next()
	}
}

// OptionalAuthMiddleware allows unauthenticated requests but sets user locals if a
// valid token is present.
func OptionalAuthMiddleware(app *core.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Next()
		}
		secret := app.Get("jwt_secret")
		if secret == nil {
			return c.Next()
		}
		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			return []byte(secret.(string)), nil
		})
		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Locals("user", claims)
				if id, ok := claims["id"].(float64); ok {
					c.Locals("user_id", uint(id))
				}
			}
		}
		return c.Next()
	}
}

// AdminMiddleware ensures the authenticated user has role "admin".
//
// BUG FIX: Original did:
//
//	claims := user.(map[string]interface{})
//
// But c.Locals("user") stores jwt.MapClaims, not map[string]interface{}.
// Although jwt.MapClaims is defined as map[string]interface{}, Go's type system
// requires an explicit assertion to jwt.MapClaims — asserting to
// map[string]interface{} panics at runtime with "interface conversion" error.
func AdminMiddleware(c *fiber.Ctx) error {
	raw := c.Locals("user")
	if raw == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// BUG FIX: assert to jwt.MapClaims, not map[string]interface{}
	claims, ok := raw.(jwt.MapClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	role, _ := claims["role"].(string)
	if role != "admin" {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden — admin access required"})
	}
	return c.Next()
}

// RequireRole returns middleware that enforces a specific role.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Locals("user")
		if raw == nil {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		claims, ok := raw.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid token claims"})
		}
		userRole, _ := claims["role"].(string)
		if userRole != role {
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
		}
		return c.Next()
	}
}

// ValidateJSONMiddleware rejects non-JSON bodies on POST/PUT/PATCH.
func ValidateJSONMiddleware(c *fiber.Ctx) error {
	method := c.Method()
	if method == "POST" || method == "PUT" || method == "PATCH" {
		ct := c.Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			return c.Status(400).JSON(fiber.Map{"error": "Content-Type must be application/json"})
		}
	}
	return c.Next()
}

// LogMiddleware logs each request method and path.
func LogMiddleware(app *core.App) fiber.Handler {
	return func(c *fiber.Ctx) error {
		app.Logger.Info("request", map[string]interface{}{
			"method": c.Method(),
			"path":   c.Path(),
			"ip":     c.IP(),
		})
		return c.Next()
	}
}
