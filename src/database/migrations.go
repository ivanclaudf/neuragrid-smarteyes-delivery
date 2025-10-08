package database

import (
	"delivery/helper"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"delivery/database/migrations"

	"gorm.io/gorm"
)

type Migration struct {
	Version string
	Migrate func(*gorm.DB) error
}

var migrationsList = make(map[string]Migration)

// Automatically get the migration function based on version
func getMigrationFunction(version string) (func(*gorm.DB) error, bool) {
	versionDigits := strings.ReplaceAll(version, ".", "")

	// Look up the function in the registry
	migrateFn, exists := migrations.MigrationRegistry[versionDigits]
	return migrateFn, exists
}

// LoadMigration loads a single migration file
func LoadMigration(filePath string) error {
	// First verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("migration file does not exist: %s", filePath)
	}

	baseName := filepath.Base(filePath)
	re := regexp.MustCompile(`^updates-([0-9]+\.[0-9]+\.[0-9]+)\.go$`)
	matches := re.FindStringSubmatch(baseName)
	if len(matches) != 2 {
		return fmt.Errorf("invalid migration filename format: %s", baseName)
	}

	version := matches[1]

	// Check if migration function exists
	migrateFn, exists := getMigrationFunction(version)
	if !exists {
		return fmt.Errorf("no migration function found for version %s", version)
	}

	// Add to migrations list
	migrationsList[version] = Migration{
		Version: version,
		Migrate: migrateFn,
	}

	return nil
}

// LoadMigrationsFromDir loads all migration files from a directory
func LoadMigrationsFromDir(dir string) error {
	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		helper.Log.Warnf("Migrations directory does not exist: %s", dir)
		return nil
	}

	// Find all migration files
	filePattern := filepath.Join(dir, "updates-*.go")
	files, err := filepath.Glob(filePattern)
	if err != nil {
		return fmt.Errorf("failed to scan migrations directory: %w", err)
	}

	// Load each migration file
	for _, file := range files {
		if err := LoadMigration(file); err != nil {
			helper.Log.Warnf("Failed to load migration file %s: %v", file, err)
		}
	}

	return nil
}

// getSortedVersions returns migration versions sorted in semantic order
func getSortedVersions() []string {
	versions := make([]string, 0, len(migrationsList))
	for v := range migrationsList {
		versions = append(versions, v)
	}

	// Sort versions
	sort.Slice(versions, func(i, j int) bool {
		// Split into components and compare numerically
		v1Parts := strings.Split(versions[i], ".")
		v2Parts := strings.Split(versions[j], ".")

		for k := 0; k < len(v1Parts) && k < len(v2Parts); k++ {
			if v1Parts[k] != v2Parts[k] {
				// Return true if v1 < v2
				return v1Parts[k] < v2Parts[k]
			}
		}

		// If all compared components are equal, the shorter one is smaller
		return len(v1Parts) < len(v2Parts)
	})

	return versions
}
