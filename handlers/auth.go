package handlers

import (
	"api/models"
	"api/utils"

	// "database/sql"

	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecretKey = []byte("supersecretkey")

// func RegisterUser(db *gorm.DB) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		var req models.RegisterRequest

// 		if err := c.Bind(&req); err != nil {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
// 		}

// 		// check if user already exists

// 		var existinguser models.User
// 		if err := db.Where("email = ?", req.Email).First(&existinguser).Error; err == nil {
// 			return c.JSON(http.StatusConflict, echo.Map{"error": "Email already exists"})
// 		}
// 		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to hash password"})
// 		}
// 		user := models.User{
// 			UserName:    req.UserName,
// 			Email:       req.Email,
// 			PhoneNumber: req.PhoneNumber,
// 			Password:    string(hash),
// 		}

// 		if err := db.Create(&user).Error; err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to create user", "details": err.Error()})
// 		}

// 		// err := db.QueryRow("INSERT INTO users (name, email, password) VALUES  ($1, $2, $3) RETURNING id, name, email", req.Name, req.Email, string(hash)).Scan(&user.ID, &user.Name, &user.Email)

// 		return c.JSON(http.StatusCreated, echo.Map{
// 			"data": map[string]string{
// 				"name":         user.UserName,
// 				"email":        user.Email,
// 				"phone_number": user.PhoneNumber,
// 			},
// 		})
// 	}

// }

func Register(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req models.RegisterRequest

		if err := c.Bind(&req); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid input", err)
		}

		if !models.IsValidRole(req.Role) || req.Role == "" {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid role. Allowed - ADMIN, USER", nil)
		}
		// check if user already exists
		var existinguser models.User

		if err := db.Where("user_name = ?", req.UserName).First(&existinguser).Error; err == nil {
			return utils.ResponseError(c, http.StatusConflict, "Username already exists", nil)
		}

		if err := db.Where("email = ?", req.Email).First(&existinguser).Error; err == nil {
			return utils.ResponseError(c, http.StatusConflict, "Email already exists", nil)
		}
		if err := db.Where("phone_number = ?", req.PhoneNumber).First(&existinguser).Error; err == nil {
			return utils.ResponseError(c, http.StatusConflict, "Phone number has being used", nil)
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to hash password", err)
		}

		user := models.User{
			UserName:    req.UserName,
			Email:       req.Email,
			Role:        req.Role,
			PhoneNumber: req.PhoneNumber,
			Password:    string(hash),
		}

		if err := db.Create(&user).Error; err != nil {
			return utils.ResponseError(c, http.StatusInternalServerError, "Failed to create user", err)
		}

		return utils.ResponseSucess(c, http.StatusCreated, "Registered successfully", map[string]string{
			"user_name":    user.UserName,
			"email":        user.Email,
			"phone_number": user.PhoneNumber,
			"role":         string(user.Role),
		})
	}
}

func Login(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req models.LoginRequest

		if err := c.Bind(&req); err != nil {
			return utils.ResponseError(c, http.StatusBadRequest, "Invalid input", err)
		}

		// check if user email exist in database

		var user models.User

		if err := db.Where("email =?", req.Email).First(&user).Error; err != nil {
			// c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid login credentials"})
			return utils.ResponseError(c, http.StatusUnauthorized, "Invalid request body", err)
		}
		// check if password matches existing one in database
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			// return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid login credentials"})
			return utils.ResponseError(c, http.StatusUnauthorized, "Invalid login credentials", err)
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID.String(),
			"exp":     time.Now().Add(24 * time.Hour).Unix(), // 24 hours
		})

		tokenString, _ := token.SignedString(jwtSecretKey)

		// return c.JSON(http.StatusOK, echo.Map{"message": "Login succesfully", "token": tokenString, "user": map[string]string{
		// 	"user_name":    user.UserName,
		// 	"email":        user.Email,
		// 	"phone_number": user.PhoneNumber,
		// 	"role":         string(user.Role),
		// }})
		return utils.ResponseSucess(c, http.StatusOK, "Login successfully", echo.Map{
			"token": tokenString,
			"user": map[string]string{
				"user_name":    user.UserName,
				"email":        user.Email,
				"phone_number": user.PhoneNumber,
				"role":         string(user.Role),
			},
		})
	}

}

// func Login(db *gorm.DB) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		var req models.LoginRequest

// 		if err := c.Bind(&req); err != nil {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
// 		}

// 		var user models.User

// 		// err := db.QueryRow("SELECT id, name, email, password FROM users WHERE email = $1", req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
// 		// check if email matches existing one in database
// 		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
// 			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid credentials"})
// 		}

// 		// check if password matches existing one in database
// 		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
// 			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid credentials"})
// 		}
// 		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 			"user_id": user.ID,
// 			"exp":     time.Now().Add(24 * time.Hour).Unix(), // 24 hours
// 		})

// 		tokenStr, _ := token.SignedString(jwtSecretKey)

// 		return c.JSON(http.StatusOK, echo.Map{
// 			"message": "Login successfully",
// 			"token":   tokenStr,
// 			"user": map[string]string{
// 				"user_name":    user.UserName,
// 				"email":        user.Email,
// 				"phone_number": user.PhoneNumber,
// 			},
// 		})

// 	}

// }
