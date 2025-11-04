package db

import (
	"api/models"
	"fmt"
	"log"
	"os"

	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDb() {
	dbConnUrl := os.Getenv("DATABASE_URL")
	if dbConnUrl == "" {
		log.Fatal("❌ DATABASE_URL not set — please configure it in your environment")
	}

	fmt.Println("🧩 Connecting to database...")

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(dbConnUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Test the connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get DB from GORM: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	fmt.Println("✅ Connected to database successfully")

	// Auto-migrate your models
	if err := db.AutoMigrate(&models.User{}, &models.Products{}, &models.StoreSetting{}); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	fmt.Println("🧱 Database migration completed successfully ✅")
}
