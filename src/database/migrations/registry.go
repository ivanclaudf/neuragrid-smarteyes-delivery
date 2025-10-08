package migrations

import (
	"gorm.io/gorm"
)

// MigrationRegistry stores all migration functions
// This allows migrations to register themselves without modifying other files
var MigrationRegistry = make(map[string]func(*gorm.DB) error)

// RegisterMigration allows a migration to register itself with the registry
func RegisterMigration(version string, migrationFn func(*gorm.DB) error) {
	versionDigits := version
	MigrationRegistry[versionDigits] = migrationFn
}
