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
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		fmt.Printf("Attempting to connect to database... (attempt %d/%d)\n", i+1, maxRetries)

		db, err := gorm.Open(postgres.New(postgres.Config{
			DSN: dbConnUrl,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})

		if err == nil {
			// Test the connection
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					// Configure connection pool
					sqlDB.SetMaxIdleConns(10)
					sqlDB.SetMaxOpenConns(100)
					sqlDB.SetConnMaxLifetime(time.Hour)

					DB = db
					fmt.Println("✅ Database connected successfully")

					// Run this once in your main.go or a migration script
					// db.Migrator().DropTable(&models.User{}, &models.Products{})
					db.AutoMigrate(&models.User{}, &models.Products{})

					return
				}
			}
		}

		fmt.Printf("Failed to connect to database: %v\n", err)

		if i < maxRetries-1 {
			fmt.Printf("Retrying in %v...\n", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	log.Fatal("Failed to connect to database after all retries")
}
