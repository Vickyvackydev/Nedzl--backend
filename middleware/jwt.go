package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

var jwtSecretKey = []byte("supersecretkey")

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid or expired token"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid or expired token"})
		}

		claims := token.Claims.(jwt.MapClaims)
		if userID, ok := claims["user_id"].(float64); ok {
			c.Set("user_id", uint(userID))
		} else {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token claims"})
		}

		return next(c)

	}

}
