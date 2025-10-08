#!/bin/bash

echo "Updating database permissions for delivery service..."

# Update the values below with your actual connection details
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="delivery"
DB_ADMIN_USER="postgres"
DB_ADMIN_PASSWORD=""  # Fill in if needed

# Database users
READER_USER="delivery_reader"
WRITER_USER="delivery_writer"

# Create temporary SQL file
SQL_FILE="/tmp/update_permissions.sql"

cat > $SQL_FILE << EOF
-- Grant permissions to reader user
GRANT CONNECT ON DATABASE ${DB_NAME} TO ${READER_USER};
GRANT USAGE ON SCHEMA public TO ${READER_USER};
GRANT SELECT ON ALL TABLES IN SCHEMA public TO ${READER_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO ${READER_USER};

-- Grant permissions to writer user
GRANT CONNECT ON DATABASE ${DB_NAME} TO ${WRITER_USER};
GRANT USAGE ON SCHEMA public TO ${WRITER_USER};
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO ${WRITER_USER};
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO ${WRITER_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO ${WRITER_USER};
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO ${WRITER_USER};

-- Ensure permissions are applied to all existing tables and sequences
-- This is critical if tables were created after the users
DO \$\$
DECLARE
    t_name text;
BEGIN
    FOR t_name IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public')
    LOOP
        EXECUTE format('GRANT SELECT ON %I TO ${READER_USER}', t_name);
        EXECUTE format('GRANT ALL PRIVILEGES ON %I TO ${WRITER_USER}', t_name);
    END LOOP;
    
    FOR t_name IN (SELECT sequencename FROM pg_sequences WHERE schemaname = 'public')
    LOOP
        EXECUTE format('GRANT USAGE, SELECT ON SEQUENCE %I TO ${READER_USER}', t_name);
        EXECUTE format('GRANT ALL PRIVILEGES ON SEQUENCE %I TO ${WRITER_USER}', t_name);
    END LOOP;
END
\$\$;
EOF

# If Docker is used (uncomment and adjust as needed)
# echo "Executing SQL script in Docker container..."
cat $SQL_FILE | docker exec -i rules2-timescaledb-1 psql -U ${DB_ADMIN_USER} -d ${DB_NAME}

# For direct connection to PostgreSQL
echo "Executing SQL script on PostgreSQL..."
export PGPASSWORD=${DB_ADMIN_PASSWORD}
#psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_ADMIN_USER} -d ${DB_NAME} -f ${SQL_FILE}

echo "Database permissions updated successfully!"

# Clean up
rm $SQL_FILE
