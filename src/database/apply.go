package database

import (
	"delivery/helper"
	"fmt"
	"time"

	"delivery/models"

	"gorm.io/gorm"
)

// ApplyDatabaseUpdates sets up and applies all database migrations
func ApplyDatabaseUpdates(db *gorm.DB, migrationsDir string) error {
	// Load and apply migrations
	if err := LoadMigrationsFromDir(migrationsDir); err != nil {
		helper.Log.Errorf("failed to load migrations: %v", err)
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	helper.Log.Info("Applying database migrations...")
	if err := ApplyMigrations(db); err != nil {
		helper.Log.Errorf("failed to apply migrations: %v", err)
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	helper.Log.Info("Database migrations completed successfully")

	return nil
}

// ApplyMigrations runs all pending migrations in order
func ApplyMigrations(db *gorm.DB) error {
	// Ensure migrations table exists
	if err := db.AutoMigrate(&models.AppUpdate{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all applied migrations
	var appliedMigrations []models.AppUpdate
	if err := db.Find(&appliedMigrations).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Create a map of applied migrations for quick lookup
	appliedMap := make(map[string]bool)
	for _, m := range appliedMigrations {
		appliedMap[m.Version] = true
	}

	// Apply pending migrations in order
	for _, version := range getSortedVersions() {
		// Skip already applied migrations
		if appliedMap[version] {
			helper.Log.Infof("Migration %s already applied, skipping", version)
			continue
		}

		helper.Log.Infof("Applying migration version %s", version)

		// Get migration function
		migrateFn, exists := getMigrationFunction(version)
		if !exists {
			helper.Log.Warnf("No migration function found for version %s, skipping", version)
			continue
		}

		// Apply migration within transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			// Run the migration
			if err := migrateFn(tx); err != nil {
				helper.Log.Errorf("Migration %s failed: %v", version, err)
				return fmt.Errorf("migration failed: %w", err)
			}

			// Record the migration as applied
			migration := models.AppUpdate{
				Version:   version,
				Applied:   true,
				AppliedAt: time.Now(),
			}

			if err := tx.Create(&migration).Error; err != nil {
				helper.Log.Errorf("Failed to record migration %s: %v", version, err)
				return fmt.Errorf("failed to record migration: %w", err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", version, err)
		}

		helper.Log.Infof("Successfully applied migration %s", version)
	}

	return nil
}
