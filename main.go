package main

import (
	"api/db"
	"api/handlers"

	jwtMiddleware "api/middleware"
	"os"
	"strconv"

	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/lib/pq"
)

func main() {

	// 🧩 Load .env file first
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	} else {
		log.Println("✅ .env file loaded successfully")
	}
	db.ConnectDb()

	// Echo instance

	// if err := migrations.MigrateToUUID(db.DB); err != nil {
	// 	log.Fatal("❌ UUID Migration failed:", err)
	// }

	// log.Println("✅ Database UUID migration completed!")
	e := echo.New()

	// Global middleware to return JSON
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS configuration for production deployment
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5175", "http://localhost:5174", "http://localhost:5173", "https://nedzl-market.vercel.app"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			return next(c)
		}
	})

	// Routes
	e.POST("/auth/register", handlers.Register(db.DB))
	e.POST("/auth/login", handlers.Login(db.DB))

	auth := e.Group("")
	auth.Use(jwtMiddleware.AuthMiddleware)

	// -- PRODUCTS ROUTES -- >
	auth.POST("/products", handlers.CreateProduct(db.DB))
	e.GET("/products", handlers.GetAllProducts(db.DB))
	e.GET("/products/:id", handlers.GetSingleProduct(db.DB))
	e.GET("/products/counts", handlers.GetTotalProductsByCatgory(db.DB))
	e.GET("/store-settings/:id", handlers.GetStoreSettings(db.DB))
	e.GET("/products/search", handlers.SearchProducts(db.DB))
	e.GET("/products/search/results", handlers.GetSearchResults(db.DB))
	auth.PUT("/products/:id/user", handlers.UpdateUserProduct(db.DB))
	auth.GET("/products/user", handlers.GetUserProducts(db.DB))
	auth.GET("/products/:id/user", handlers.GetUserProduct(db.DB))
	auth.DELETE("/products/:id/user", handlers.DeleteUserProduct(db.DB))
	auth.DELETE("/products/:id", handlers.DeleteUserProduct(db.DB))
	auth.PATCH("/products/update/:id/status", handlers.UpdateProductStatus(db.DB))

	// -- USER ROUTES -->
	auth.GET("/me", handlers.Me)
	auth.PATCH("/users/update", handlers.UpdateUser(db.DB))
	auth.PATCH("/users/update/:id/status", handlers.UpdateUserStatus(db.DB))
	auth.POST("/store-settings", handlers.CreateStoreSettings(db.DB))

	auth.GET("/users", handlers.GetUsers(db.DB))
	auth.GET("/users/:id", handlers.GetUserById(db.DB))

	// -- ADMIN ROUTES -->
	auth.GET("/admin/overview", handlers.GetDashboardOverview(db.DB))
	auth.GET("/admin/user/overview", handlers.GetUserDashboardOverview(db.DB))
	auth.GET("/admin/users", handlers.GetDashboardUsers(db.DB))
	auth.GET("/admin/products", handlers.GetAdminProducts(db.DB))
	auth.GET("/admin/user/:id", handlers.GetUserDetails(db.DB))
	auth.POST("/admin/feature-products/:box_number", handlers.UpdateFeaturedSection(db.DB))
	e.GET("/feature-products", handlers.GetFeaturedSections(db.DB))
	auth.GET("/admin/feature-products", handlers.GetFeaturedSections(db.DB))
	auth.DELETE("/admin/product/:id/delete", handlers.DeleteAdminProduct(db.DB))
	auth.DELETE("/admin/users/:id/delete", handlers.DeleteUser(db.DB))

	// Get port from environment variable (Railway provides PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default port for local development
	}

	address := "0.0.0.0:" + port
	log.Printf("Server running at http://%s", address)
	e.Logger.Fatal(e.Start(address))

}

// func createUser(db *sql.DB) echo.HandlerFunc {

// 	return func(c echo.Context) error {
// 		var u User
// 		if err := c.Bind(&u); err != nil {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
// 		}

// 		err := db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", u.Name, u.Email).Scan(&u.ID)

// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})

// 		}

// 		return c.JSON(http.StatusCreated, u)
// 	}

// }

// func getUsers(db *sql.DB) echo.HandlerFunc {

// 	return func(c echo.Context) error {
// 		rows, err := db.Query("SELECT id, name, email FROM users")
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})

// 		}

// 		defer rows.Close()

// 		var users []User

// 		for rows.Next() {
// 			var u User
// 			if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
// 				return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})

// 			}
// 			users = append(users, u)
// 		}
// 		return c.JSON(http.StatusOK, users)
// 	}

// }

// func getUser(db *sql.DB) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		id := c.Param("id")
// 		var u User

// 		err := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email)

// 		if err == sql.ErrNoRows {
// 			return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})

// 		} else if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
// 		}

// 		return c.JSON(http.StatusOK, u)
// 	}

// }

// func updateUser(db *sql.DB) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		id := c.Param("id")
// 		var u User

// 		if err := c.Bind(&u); err != nil {
// 			return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})

// 		}

// 		result, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", u.Name, u.Email, id)
// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})

// 		}

// 		rowsAffected, _ := result.RowsAffected()

// 		if rowsAffected == 0 {
// 			return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
// 		}

// 		u.ID = atoi(id)
// 		return c.JSON(http.StatusOK, u)
// 	}

// }

// func deleteUser(db *sql.DB) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		id := c.Param("id")

// 		result, err := db.Exec("DELETE FROM users WHERE id = $1", id)

// 		if err != nil {
// 			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})

// 		}

// 		rowsAffected, _ := result.RowsAffected()
// 		if rowsAffected == 0 {
// 			return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})

// 		}

// 		return c.NoContent(http.StatusNoContent)
// 	}

// }

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i

}

// func jsonContentTypeMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "application/json")
// 		next.ServeHTTP(w, r)
// 	})
// }

// func getUsers(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		rows, err := db.Query("SELECT * FROM users")
// 		if err != nil {
// 			log.Fatal(err)

// 		}

// 		defer rows.Close()

// 		users := []User{}

// 		for rows.Next() {
// 			var u User

// 			if err := rows.Scan(&u.ID, &u.Name, &u.Email); err != nil {
// 				log.Fatal()
// 			}

// 			users = append(users, u)

// 		}

// 		if err := rows.Err(); err != nil {
// 			log.Fatal(err)
// 		}

// 		json.NewEncoder(w).Encode(users)
// 	}

// }

// func getUser(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// vars := mux.Vars(r)

// 		// id := vars["id"]

// 		path := r.URL.Path // "/user/123"

// 		parts := strings.Split(path, "/") // ["", "user", "123"]

// 		if len(parts) != 3 || parts[2] == "" {
// 			http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 			return
// 		}

// 		id := parts[2]
// 		var u User
// 		err := db.QueryRow("SELECT * FROM users WHERE id = $1", id).Scan(&u.ID, &u.Name, &u.Email)

// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		json.NewEncoder(w).Encode(u)
// 	}
// }
// func createUser(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var u User
// 		json.NewDecoder(r.Body).Decode(&u)

// 		err := db.QueryRow("INSERT INTO users (name, email) values ($1, $2) RETURNING id", u.Name, u.Email)

// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		json.NewEncoder(w).Encode(u)
// 	}

// }
// func updateUser(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var u User
// 		json.NewDecoder(r.Body).Decode(&u)

// 		// vars := mux.Vars(r)
// 		// id := vars["id"]

// 		path := r.URL.Path // "/user/123"

// 		parts := strings.Split(path, "/") // ["", "user", "123"]

// 		if len(parts) != 3 || parts[2] == "" {
// 			http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 			return
// 		}

// 		id := parts[2]
// 		_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE ID = $3", u.Name, u.Email, id)

// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		json.NewEncoder(w).Encode(w)
// 	}
// }
// func deleteUser(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// vars := mux.Vars(r)
// 		// id := vars["id"]
// 		path := r.URL.Path // "/user/123"

// 		parts := strings.Split(path, "/") // ["", "user", "123"]

// 		if len(parts) != 3 || parts[2] == "" {
// 			http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 			return
// 		}

// 		id := parts[2]

// 		_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		json.NewEncoder(w).Encode("User deleted")
// 	}

// }
