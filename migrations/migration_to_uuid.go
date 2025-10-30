package migrations

import (
	"fmt"
	"log"

	// "github.com/google/uuid"
	"gorm.io/gorm"
)

// MigrateToUUID converts all integer primary keys to UUID in your tables.
func MigrateToUUID(db *gorm.DB) error {
	log.Println("🔄 Starting UUID migration...")

	// 1️⃣ Ensure PostgreSQL extensions exist
	extensions := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
	}

	for _, ext := range extensions {
		if err := db.Exec(ext).Error; err != nil {
			return fmt.Errorf("failed to install extension: %w", err)
		}
	}

	log.Println("✅ UUID extensions installed (uuid-ossp, pgcrypto)")

	// 2️⃣ List all tables you want to migrate
	tables := []string{"users", "products", "orders"} // Add more as needed

	for _, table := range tables {
		log.Printf("Migrating table: %s...", table)

		// 2.1 Add new UUID column temporarily
		if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN new_id UUID DEFAULT gen_random_uuid()`, table)).Error; err != nil {
			return fmt.Errorf("failed to add new_id column to %s: %w", table, err)
		}

		// 2.2 Populate new UUIDs
		if err := db.Exec(fmt.Sprintf(`UPDATE %s SET new_id = gen_random_uuid()`, table)).Error; err != nil {
			return fmt.Errorf("failed to update UUIDs for %s: %w", table, err)
		}

		// 2.3 Drop foreign key constraints referencing this table
		// NOTE: optional — if you have foreign keys to this table, they must be dropped manually or using introspection logic

		// 2.4 Replace the old PK
		if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s DROP CONSTRAINT %s_pkey`, table, table)).Error; err != nil {
			log.Printf("ℹ️ Skipping drop primary key (might already be removed): %v", err)
		}

		if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s DROP COLUMN id`, table)).Error; err != nil {
			return fmt.Errorf("failed to drop old id column for %s: %w", table, err)
		}

		if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s RENAME COLUMN new_id TO id`, table)).Error; err != nil {
			return fmt.Errorf("failed to rename new_id column for %s: %w", table, err)
		}

		if err := db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD PRIMARY KEY (id)`, table)).Error; err != nil {
			return fmt.Errorf("failed to add new primary key for %s: %w", table, err)
		}

		log.Printf("✅ Table %s migrated successfully.", table)
	}

	log.Println("🎉 UUID migration completed successfully!")
	return nil
}
