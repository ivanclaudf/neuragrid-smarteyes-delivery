# Delivery Service

A robust Golang microservice for managing multi-channel messaging delivery including Email, SMS, and WhatsApp communications. Built with enterprise-grade architecture patterns including read-write database separation, asynchronous processing, and pluggable provider system.

## Features

### Core Messaging Capabilities
- **Email delivery** with SendGrid provider support
- **SMS delivery** via Twilio provider
- **WhatsApp messaging** through Twilio Business API
- **Template management** with Go text/template variable substitution
- **Multi-tenant** support for SaaS applications

### Architecture & Infrastructure
- **Pluggable provider system** - Easy extension with new messaging providers
- **Database migrations framework** - Automated schema management
- **Read-write database separation** - Optimized for performance and scaling
- **Asynchronous message processing** with Apache Pulsar message queue
- **Secure credential storage** with encryption
- **Docker containerization** for easy deployment
- **RESTful API** with comprehensive endpoints for all operations

### Enterprise Features
- **Message tracking and events** - Full delivery lifecycle monitoring
- **Provider configuration management** - Dynamic provider setup and switching
- **Template versioning** - Code-based template management with provider-specific IDs
- **Tenant isolation** - Complete multi-tenant data separation
- **Comprehensive logging** - Structured logging for monitoring and debugging

## Project Architecture

```
├── src/
│   ├── main.go                    # Application entry point
│   ├── api/                       # Business logic layer
│   │   ├── email.go              # Email API operations
│   │   ├── sms.go                # SMS API operations  
│   │   ├── whatsapp.go           # WhatsApp API operations
│   │   ├── template.go           # Template management
│   │   └── provider.go           # Provider configuration
│   ├── handler/                   # HTTP handlers (presentation layer)
│   ├── models/                    # Data models and database entities
│   ├── services/                  # External service integrations
│   │   ├── providers/            # Messaging provider implementations
│   │   └── queue/                # Pulsar queue consumer/producer
│   ├── database/                  # Database connection and migrations
│   └── helper/                    # Utilities and common functions
├── docs/                          # Comprehensive documentation
└── docker-compose.yml             # Local development setup
```

## Documentation

The project documentation is comprehensively organized:

- **[API Documentation](docs/api.md)** - Complete REST API reference with request/response examples
- **[Database Documentation](docs/database.md)** - Schema design, setup instructions, and migration guide  
- **[Template Documentation](docs/templates.md)** - Template creation and variable substitution guide
- **[Direct Pulsar Integration](docs/direct_pulsar_integration.md)** - Message queue integration for advanced use cases

## Quick Start

### Prerequisites

- **Docker & Docker Compose** - For containerized setup
- **PostgreSQL 14+** - Database backend  
- **Apache Pulsar** - Message queue (included in docker-compose)
- **Go 1.23+** - For local development

### Environment Configuration

Configure the following environment variables for your deployment:

#### Database Settings
```bash
# Writer database (for write operations)
DB_HOST=localhost
DB_PORT=5432
DB_USER=delivery_writer
DB_PASSWORD=writer_password
DB_NAME=delivery

# Reader database (for read operations - can be same as writer)
DB_READER_HOST=localhost
DB_READER_PORT=5432
DB_READER_USER=delivery_reader
DB_READER_PASSWORD=reader_password
DB_READER_NAME=delivery
```

#### Application Configuration
```bash
VERSION=0.1.0
LOG_LEVEL=info
PULSAR_URL=pulsar://localhost:6650
ENCRYPTION_KEY=32_character_encryption_key_here
```

#### Provider Credentials (Configure as needed)
```bash
# SendGrid for Email
SENDGRID_API_KEY=your_sendgrid_api_key
SENDGRID_FROM_EMAIL=your_from_email

# Twilio for SMS & WhatsApp
TWILIO_ACCOUNT_SID=your_twilio_account_sid
TWILIO_AUTH_TOKEN=your_twilio_auth_token
TWILIO_FROM_NUMBER=your_twilio_from_number
TWILIO_WHATSAPP_FROM=whatsapp:your_twilio_whatsapp_number
```

### Running with Docker Compose

1. **Clone the repository**
```bash
git clone <repository-url>
cd delivery
```

2. **Start all services**
```bash
docker-compose up -d
```

This will start:
- PostgreSQL database 
- Apache Pulsar message queue
- The delivery service API (available on port 8060)

### Running for Development

1. **Set up environment variables** (see configuration above)

2. **Start dependencies**
```bash
# Start only database and Pulsar
docker-compose up -d db pulsar
```

3. **Run the application**
```bash
cd src
go run main.go
```

The API will be available at `http://localhost:8080`

## Database Setup

The service uses PostgreSQL with a read-write separation pattern. For detailed setup instructions including user creation, permissions, and migration information, refer to the [Database Documentation](docs/database.md).

## API Usage

The service provides REST APIs for all messaging operations. Here are the main endpoints:

- **Health Check**: `GET /api/v1/health` - Service status
- **Email**: `POST /api/v1/email` - Send email messages  
- **SMS**: `POST /api/v1/sms` - Send SMS messages
- **WhatsApp**: `POST /api/v1/whatsapp` - Send WhatsApp messages
- **Templates**: `POST/GET/PUT /api/v1/templates` - Manage message templates
- **Providers**: `POST/GET/PUT /api/v1/providers` - Configure messaging providers

All messages are processed asynchronously through Apache Pulsar queues. For complete API documentation with request/response examples, see the [API Documentation](docs/api.md).

## Key Concepts

### Providers
Messaging providers (like Twilio, SendGrid) are configurable entities that handle actual message delivery. Each provider has:
- **Configuration** - Public settings (URLs, account IDs)
- **Secure Configuration** - Encrypted sensitive data (API keys, tokens)  
- **Channel Support** - EMAIL, SMS, or WHATSAPP
- **Tenant Isolation** - Provider configs are tenant-specific

### Templates
Templates enable dynamic message content with variable substitution:
- **Go text/template syntax** - `{{.variable}}` placeholders
- **Multi-channel support** - Same template for different channels
- **Provider-specific IDs** - Link to provider template systems
- **Version control** - Templates are immutable after creation

### Message Processing
1. **API Request** - Messages submitted via REST endpoints
2. **Queue Publishing** - Messages sent to Pulsar topics  
3. **Consumer Processing** - Background workers process queued messages
4. **Provider Delivery** - Actual delivery via configured providers
5. **Event Tracking** - Delivery status updates and events

## Extending the System

### Adding New Providers

1. **Implement the service interface** (`EmailService`, `SMSService`, or `WhatsAppService`)
2. **Add to the factory** (in `services/providers/*_factory.go`)
3. **Configure provider** via the Provider API

### Adding New Message Types

1. **Define models** in `models/` package
2. **Create API handlers** in `api/` and `handler/` packages  
3. **Implement queue consumers** in `services/queue/`
4. **Add database migrations** if needed

## Contributing

1. **Follow Go conventions** - Use `gofmt` and standard project layout
2. **Add tests** - Unit tests for business logic, integration tests for APIs
3. **Update documentation** - Keep API docs and examples current
4. **Database changes** - Always include migrations for schema changes

For more details on development setup and guidelines, see [CONTRIBUTING.md](CONTRIBUTING.md).

---

**Need help?** Check the documentation in the `docs/` folder or review the API documentation for detailed examples.
