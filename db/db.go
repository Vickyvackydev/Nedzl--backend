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

	logLevel := logger.Info
	if os.Getenv("ENV") != "development" {
		logLevel = logger.Error
	}

	db, err := gorm.Open(postgres.Open(dbConnUrl), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ Failed to get DB from GORM: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db

	fmt.Println("🧱 Running database extension setup...")

	// ✅ Enable UUID extension safely
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`)
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`) // optional but useful
	// db.Exec(`ALTER TABLE customer_reviews
	// DROP CONSTRAINT IF EXISTS fk_customer_reviews_product_details;`)

	fmt.Println("🧱 Extensions ready")

	// Auto-migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Products{},
		&models.StoreSetting{},
		&models.FeaturedSection{},
		&models.FeaturedSectionProduct{},
		&models.CustomerReview{},
		&models.ProductLike{},
		&models.Contact{},
	); err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	fmt.Println("✅ Database migration completed")
}
