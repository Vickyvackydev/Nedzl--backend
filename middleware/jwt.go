package middleware

import (
	"net/http"
	"strings"

	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("supersecretkey") // Fallback for dev, but should be set in production
	}
	return []byte(secret)
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid or expired token"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return getJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid or expired token"})
		}

		claims := token.Claims.(jwt.MapClaims)
		if userIDStr, ok := claims["user_id"].(string); ok {
			uid, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token user id"})
			}
			c.Set("user_id", uid)
		} else {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token claims"})
		}

		if role, ok := claims["role"].(string); ok {
			c.Set("role", role)
		}

		return next(c)

	}

}

func OptionalAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		// If no header or invalid format, just proceed without setting user_id
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return next(c)
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return getJWTSecret(), nil
		})

		// If invalid token, just proceed (or maybe we should log it? For now, treat as guest)
		if err != nil || !token.Valid {
			return next(c)
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userIDStr, ok := claims["user_id"].(string); ok {
				if uid, err := uuid.Parse(userIDStr); err == nil {
					c.Set("user_id", uid)
				}
			}
			if role, ok := claims["role"].(string); ok {
				c.Set("role", role)
			}
		}

		return next(c)
	}
}

func IsAdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role, ok := c.Get("role").(string)
		if !ok || role != "ADMIN" {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "Admin access required"})
		}
		return next(c)
	}
}
