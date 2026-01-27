package db

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// SyncReferralCounts updates the referral_count for all users based on the referral_by field.
// This is a one-time migration script to ensure the denormalized counts are accurate.
func SyncReferralCounts(db *gorm.DB) error {
	fmt.Println("Starting referral count synchronization...")

	// Use a raw SQL query for maximum efficiency, especially for millions of users.
	// This query counts occurrences of each user's ID in the referral_by JSONB field.
	err := db.Exec(`
		UPDATE users u
		SET referral_count = (
			SELECT count(*)
			FROM users r
			WHERE (r.referral_by->>'id')::uuid = u.id
		)
	`).Error

	if err != nil {
		log.Printf("Failed to sync referral counts: %v", err)
		return err
	}

	fmt.Println("Referral counts synchronized successfully")
	return nil
}
