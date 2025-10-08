#!/bin/bash

# Script to update template IDs in the database
# This script updates template IDs for a specific template UUID

# Template UUID to update
TEMPLATE_UUID="e4632aa8-586d-4789-947c-7e71ed8be33b"

# The Twilio template ID (HX format)
TWILIO_TEMPLATE_ID="HXb5b62575e6e4ff6129ad7c8efe1f983e"

# Update the template in the database
# Replace 'psql' parameters with your actual database connection details
psql -h localhost -U your_username -d your_database -c "
UPDATE templates 
SET template_ids = jsonb_set(
    CASE 
        WHEN template_ids IS NULL THEN '{}'::jsonb 
        ELSE template_ids 
    END, 
    '{twilio}', 
    '\"$TWILIO_TEMPLATE_ID\"'::jsonb
)
WHERE uuid = '$TEMPLATE_UUID';
"

echo "Updated template ID for template UUID: $TEMPLATE_UUID"
echo "Set Twilio template ID to: $TWILIO_TEMPLATE_ID"
