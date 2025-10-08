# Delivery Service

A Golang application for managing various messaging delivery channels including Email, SMS, and WhatsApp.

## Features

- Email delivery (SendGrid provider)
- SMS delivery (Twilio provider)
- WhatsApp delivery (Twilio provider)
- Abstracted interfaces for easy extension with new providers
- Database migrations framework
- Asynchronous message processing with Apache Pulsar
- Template management with variable substitution using Go's text/template
- Secure credential storage with encryption
- Docker support

## Documentation

The project documentation is organized into several files:

- [API Documentation](docs/api.md): API endpoints, request/response formats, and examples
- [Database Documentation](docs/database.md): Database schema, setup, and migrations
- [Template Documentation](docs/templates.md): Guide on creating and using templates with variable support

## Setup

### Environment Variables

The following environment variables need to be set:

```
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=delivery_writer
DB_PASSWORD=writer_password
DB_NAME=delivery
DB_READER_HOST=localhost
DB_READER_PORT=5432
DB_READER_USER=delivery_reader
DB_READER_PASSWORD=reader_password
DB_READER_NAME=delivery

# Application settings
VERSION=0.1.0
LOG_LEVEL=info

# Email provider settings
SENDGRID_API_KEY=your_sendgrid_api_key
SENDGRID_FROM_EMAIL=your_from_email

# SMS provider settings
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_twilio_auth_token
TWILIO_FROM_NUMBER=your_twilio_from_number

# WhatsApp provider settings
TWILIO_WHATSAPP_FROM=whatsapp:your_twilio_whatsapp_number

# Pulsar settings
PULSAR_URL=pulsar://localhost:6650

# Security
ENCRYPTION_KEY=32_character_encryption_key_here
```

### Database Setup

For detailed database setup instructions, including creating read/write users and understanding migrations, see the [Database Documentation](docs/database.md).

### Running the Application

1. Clone this repository
2. Configure environment variables in docker-compose.yml
3. Run with Docker Compose:

```bash
docker-compose up -d
```

## Development

To run locally:

1. Set up your environment variables
2. Run the application:

```bash
cd src
go run main.go
```

## API Overview

### WhatsApp API

The WhatsApp API allows sending template messages to recipients through WhatsApp. Messages are processed asynchronously.

```
POST /api/v2/whatsapp
```

Request body:

```json
{
  "messages": [
    {
      "template": "421bb248904716d53b9b56ce43a0f24c",
      "to": [
        {
          "name": "Ivan Claud",
          "telephone": "+9900186039"
        }
      ],
      "provider": "0bca5714-bceb-49a4-a4eb-e3afcec26328",
      "refno": "000000000001",
      "categories": [
        "detection_alerts"
      ],
      "identifiers": {
        "tenant": "example-tenant",
        "eventUuid": "0bca5714-bceb-49a4-a4eb-e3afcec26328",
        "actionUuid": "0bca5714-bceb-49a4-a4eb-e3afcec26328",
        "actionCode": "notify_supervisor"
      },
      "params": {
        "name": "Ivan"
      },
      "attachments": {
        "inline": [
          {
            "filename": "logo.png",
            "type": "image/png",
            "content": "base64 encoded file content",
            "contentId": "logo"
          }
        ]
      }
    }
  ]
}
```

Response:

```json
{
  "code": "success",
  "message": "WhatsApp messages accepted for processing",
  "data": {
    "messages": [
      {
        "refno": "000000000001",
        "uuid": "a1b2c3d4-e5f6-7890-abcd-1234567890ab"
      }
    ]
  }
}
```

For more detailed API documentation, see the [API Documentation](docs/api.md) file.
