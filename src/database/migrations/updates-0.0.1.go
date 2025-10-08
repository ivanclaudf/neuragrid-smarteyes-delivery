package migrations

import (
	"delivery/models"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	RegisterMigration("001", ApplyMigrationV001)
}

// ApplyMigrationV001 initializes the database with necessary tables
func ApplyMigrationV001(db *gorm.DB) error {
	// UUIDs are now generated in application code using the helper.GenerateUUID function

	// Create AppUpdate table (if it doesn't exist already)
	if err := db.AutoMigrate(&models.AppUpdate{}); err != nil {
		return fmt.Errorf("failed to create app_updates table: %v", err)
	}

	// Create messages table
	if err := db.AutoMigrate(&models.Message{}); err != nil {
		return fmt.Errorf("failed to create messages table: %v", err)
	}

	// Create GIN index for JSON fields in messages table
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_identifiers ON messages USING GIN (identifiers jsonb_path_ops)").Error; err != nil {
		return fmt.Errorf("failed to create GIN index on messages.identifiers: %v", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_messages_categories ON messages USING GIN (categories jsonb_path_ops)").Error; err != nil {
		return fmt.Errorf("failed to create GIN index on messages.categories: %v", err)
	}

	// Create templates table
	if err := db.AutoMigrate(&models.Template{}); err != nil {
		return fmt.Errorf("failed to create templates table: %v", err)
	}

	// Create providers table
	if err := db.AutoMigrate(&models.Provider{}); err != nil {
		return fmt.Errorf("failed to create providers table: %v", err)
	}

	// Create GIN index for JSON fields in providers table
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_providers_config ON providers USING GIN (config jsonb_path_ops)").Error; err != nil {
		return fmt.Errorf("failed to create GIN index on providers.config: %v", err)
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_providers_secure_config ON providers USING GIN (secure_config jsonb_path_ops)").Error; err != nil {
		return fmt.Errorf("failed to create GIN index on providers.secure_config: %v", err)
	}

	// Create message_events table
	if err := db.AutoMigrate(&models.MessageEvent{}); err != nil {
		return fmt.Errorf("failed to create message_events table: %v", err)
	}

	// Create GIN index for JSON field in message_events table
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_message_events_metadata ON message_events USING GIN (metadata jsonb_path_ops)").Error; err != nil {
		return fmt.Errorf("failed to create GIN index on message_events.metadata: %v", err)
	}

	return nil
}
