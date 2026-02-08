package db

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

func UpdateDB(db *gorm.DB) error {

	// Use a raw SQL query for maximum efficiency, especially for millions of users.
	// This query counts occurrences of each user's ID in the referral_by JSONB field.
	err := db.Exec(`
		UPDATE products
SET name = product_name
WHERE name IS NULL OR name = '';
	`).Error

	if err != nil {
		log.Printf("Error occured while updating the database: %v", err)
		return err
	}

	fmt.Println("Error occured while updating the database")
	return nil
}
