package database

import (
	"fmt"
	"os"
	"time"

	"delivery/helper"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBConnection stores both writer and reader database connections
type DBConnection struct {
	Writer *gorm.DB
	Reader *gorm.DB
}

// Connect establishes connections to both writer and reader databases
// with retry logic
func Connect() (*DBConnection, error) {
	helper.Log.Info("Initializing database connections...")

	// Set up writer database connection
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Initialize reader database connection
	readerDsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_READER_HOST"),
		os.Getenv("DB_READER_PORT"),
		os.Getenv("DB_READER_USER"),
		os.Getenv("DB_READER_PASSWORD"),
		os.Getenv("DB_READER_NAME"),
	)

	var db *gorm.DB
	var readerDB *gorm.DB
	var err error

	maxRetries := 10

	// Connect to writer DB with retry logic
	for i := 0; i < maxRetries; i++ {
		helper.Log.Infof("Attempting writer database connection (attempt %d of %d)...", i+1, maxRetries)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			helper.Log.Infof("Successfully connected to writer database after %d attempt(s)", i+1)
			break
		}
		if i < maxRetries-1 {
			retryDelay := time.Duration(i+1) * time.Second
			helper.Log.Warnf("Failed to connect to writer database, retrying in %v... (attempt %d/%d)", retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
		} else {
			helper.Log.Errorf("Failed to connect to writer database after %d attempts: %v", maxRetries, err)
			return nil, err
		}
	}

	// Connect to reader DB with retry logic
	for i := 0; i < maxRetries; i++ {
		helper.Log.Infof("Attempting reader database connection (attempt %d of %d)...", i+1, maxRetries)
		readerDB, err = gorm.Open(postgres.Open(readerDsn), &gorm.Config{})
		if err == nil {
			helper.Log.Infof("Successfully connected to reader database after %d attempt(s)", i+1)
			break
		}
		if i < maxRetries-1 {
			retryDelay := time.Duration(i+1) * time.Second
			helper.Log.Warnf("Failed to connect to reader database, retrying in %v... (attempt %d/%d)", retryDelay, i+1, maxRetries)
			time.Sleep(retryDelay)
		} else {
			helper.Log.Errorf("Failed to connect to reader database after %d attempts: %v", maxRetries, err)
			return nil, err
		}
	}

	return &DBConnection{
		Writer: db,
		Reader: readerDB,
	}, nil
}
