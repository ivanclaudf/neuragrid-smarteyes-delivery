# Database Documentation

This document provides details about the database structure, setup, and usage in the Delivery Service.

## Database Schema

The service uses PostgreSQL for data storage.

### Core Tables

The database includes tables for tracking message deliveries, status, templates, providers, and application updates.

#### AppUpdate

The `app_update` table tracks database migrations.

| Column    | Type         | Description                                  |
|-----------|--------------|----------------------------------------------|
| id        | serial       | Primary key                                  |
| version   | varchar(20)  | Version number (e.g., "0.0.1")               |
| applied   | boolean      | Whether the migration was successfully applied|
| applied_at| timestamp    | When the migration was applied               |
| created_at| timestamp    | When the record was created                  |
| updated_at| timestamp    | When the record was last updated             |

#### Provider

The `provider` table stores messaging service provider configurations.

| Column      | Type         | Description                                     |
|-------------|--------------|-------------------------------------------------|
| id          | serial       | Primary key                                     |
| uuid        | varchar(36)  | Unique identifier                               |
| code        | varchar(255) | Provider code (unique per tenant)               |
| provider    | varchar(255) | Provider implementation class (e.g., "TWILIO")  |
| name        | varchar(255) | Human-readable provider name                    |
| config      | jsonb        | Public provider configuration                   |
| secure_config| jsonb       | Encrypted provider configuration                |
| status      | smallint     | Provider status (0=inactive, 1=active)          |
| channel     | varchar(10)  | Message channel (WHATSAPP, SMS, EMAIL)          |
| tenant      | varchar(255) | Tenant identifier                               |
| created_at  | timestamp    | When the record was created                     |
| updated_at  | timestamp    | When the record was last updated                |

Unique index on `code` and `tenant` to ensure provider codes are unique within a tenant.

#### Template

The `template` table stores message templates.

| Column      | Type         | Description                                   |
|-------------|--------------|-----------------------------------------------|
| id          | serial       | Primary key                                   |
| uuid        | varchar(36)  | Unique identifier                             |
| name        | varchar(255) | Template name                                 |
| content     | text         | Template content with placeholders            |
| status      | smallint     | Template status (0=inactive, 1=active)        |
| channel     | varchar(10)  | Message channel (WHATSAPP, SMS, EMAIL)        |
| tenant      | varchar(255) | Tenant identifier                             |
| created_at  | timestamp    | When the record was created                   |
| updated_at  | timestamp    | When the record was last updated              |

#### Message

The `message` table tracks message deliveries.

| Column      | Type         | Description                                   |
|-------------|--------------|-----------------------------------------------|
| id          | serial       | Primary key                                   |
| uuid        | varchar(36)  | Unique identifier                             |
| refno       | varchar(255) | Reference number for tracking                 |
| template_id | varchar(36)  | Template UUID                                 |
| provider_id | varchar(36)  | Provider UUID                                 |
| recipient   | varchar(255) | Recipient identifier (phone, email)           |
| params      | jsonb        | Template parameters                           |
| status      | smallint     | Delivery status                               |
| channel     | varchar(10)  | Message channel (WHATSAPP, SMS, EMAIL)        |
| tenant      | varchar(255) | Tenant identifier                             |
| categories  | text[]       | Message categories                            |
| created_at  | timestamp    | When the record was created                   |
| updated_at  | timestamp    | When the record was last updated              |

#### MessageEvent

The `message_event` table tracks events related to message deliveries.

| Column      | Type         | Description                                   |
|-------------|--------------|-----------------------------------------------|
| id          | serial       | Primary key                                   |
| uuid        | varchar(36)  | Unique identifier                             |
| message_id  | varchar(36)  | Message UUID                                  |
| event_type  | varchar(50)  | Event type (sent, delivered, read, etc.)      |
| provider_ref| varchar(255) | Provider reference ID                         |
| data        | jsonb        | Event data                                    |
| created_at  | timestamp    | When the record was created                   |
| updated_at  | timestamp    | When the record was last updated              |

## Database Setup

### Prerequisites

- PostgreSQL 14+ running in Docker or locally

### User Configuration

Before running the application, you need to set up the database users:

```bash
# Connect to PostgreSQL
docker exec -it delivery-db-1 psql -U postgres

# Create database
postgres=# create database delivery;
CREATE DATABASE

# Connect to the new database
postgres=# \c delivery
You are now connected to database "delivery" as user "postgres".

# Create reader user
delivery=# CREATE USER delivery_reader WITH PASSWORD 'reader_password';
CREATE ROLE
delivery=# GRANT CONNECT ON DATABASE delivery TO delivery_reader;
GRANT
delivery=# GRANT USAGE ON SCHEMA public TO delivery_reader;
GRANT
delivery=# GRANT SELECT ON ALL TABLES IN SCHEMA public TO delivery_reader;
GRANT
delivery=# ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO delivery_reader;
ALTER DEFAULT PRIVILEGES

# Create writer user
delivery=# CREATE USER delivery_writer WITH PASSWORD 'writer_password';
CREATE ROLE
delivery=# GRANT CONNECT ON DATABASE delivery TO delivery_writer;
GRANT
delivery=# GRANT USAGE ON SCHEMA public TO delivery_writer;
GRANT
delivery=# GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO delivery_writer;
GRANT
delivery=# ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO delivery_writer;
ALTER DEFAULT PRIVILEGES

# Verify created users
postgres=# \du delivery_reader
postgres=# \du delivery_writer
```

## Database Migrations

The service uses a custom migration framework to manage database schema changes.

> **Note:** All UUIDs are generated in the application code using Go's cryptographic random number generator, eliminating the need for database-specific UUID generation extensions.

### Migration Files

Migrations are defined in Go files in the `src/database/migrations` directory. Each migration file follows the naming pattern `updates-X.Y.Z.go` (e.g., `updates-0.0.1.go`).

Example migration file:

```go
package migrations

import (
	"delivery/models"
	"gorm.io/gorm"
)

func init() {
	RegisterMigration("001", ApplyMigrationV001)
}

// ApplyMigrationV001 initializes the database with necessary tables
func ApplyMigrationV001(db *gorm.DB) error {
	// Create AppUpdate table (if it doesn't exist already)
	if err := db.AutoMigrate(&models.AppUpdate{}); err != nil {
		return err
	}
	
	return nil
}
```

### Migration Process

When the service starts:

1. The application connects to the database
2. It checks for any pending migrations
3. Migrations are applied in order based on their version numbers
4. Each applied migration is recorded in the `app_update` table

This ensures that database schema changes are applied consistently and only once.

## Read-Write Separation

The application uses a read-write separation pattern for database access:

- **Writer DB Connection**: Used for operations that modify data
- **Reader DB Connection**: Used for read-only operations

This pattern improves performance and allows for potential future scaling with read replicas.

## Environment Variables

Database connection settings are configured using the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| DB_HOST | Writer database hostname | - |
| DB_PORT | Writer database port | 5432 |
| DB_USER | Writer database username | - |
| DB_PASSWORD | Writer database password | - |
| DB_NAME | Writer database name | - |
| DB_READER_HOST | Reader database hostname | - |
| DB_READER_PORT | Reader database port | 5432 |
| DB_READER_USER | Reader database username | - |
| DB_READER_PASSWORD | Reader database password | - |
| DB_READER_NAME | Reader database name | - |
