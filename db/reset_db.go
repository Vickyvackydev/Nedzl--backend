package db

import (
	"log"

	"api/models"

	"gorm.io/gorm"
)

func ResetDatabase(db *gorm.DB) {
	log.Println("⚠️ Resetting entire database...")

	// Drop tables in dependency-safe order
	err := db.Migrator().DropTable(
		&models.FeaturedSectionProduct{},
		&models.FeaturedSection{},
		&models.CustomerReview{},
		&models.Products{},
		&models.StoreSetting{},
		&models.User{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to drop tables: %v", err)
	}

	log.Println("🧹 All tables dropped.")
	log.Println("🔄 Recreating tables...")

	// Create tables again
	err = db.AutoMigrate(
		&models.User{},
		&models.Products{},
		&models.StoreSetting{},
		&models.CustomerReview{},
		&models.FeaturedSection{},
		&models.FeaturedSectionProduct{},
	)

	if err != nil {
		log.Fatalf("❌ AutoMigrate failed: %v", err)
	}

	log.Println("✅ Database reset complete! Fresh schema is ready.")
}
