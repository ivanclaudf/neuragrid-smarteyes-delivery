# API Documentation

This document provides detailed information about the Delivery Service API endpoints, including request and response formats.

## Health Check

### `GET /api/v1/health`

Check the health status of the API.

**Response:**

```json
{
    "status": "ok",
    "version": "0.1.0"
}
```

## WhatsApp API

### `POST /api/v1/whatsapp`

Send WhatsApp template messages to recipients. Messages are queued and processed asynchronously.

**Request:**

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

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| messages | array | Yes | Array of message objects to send |
| messages[].template | string | Yes | UUID of the template to use |
| messages[].to | array | Yes | Array of recipient objects |
| messages[].to[].name | string | No | Name of the recipient |
| messages[].to[].telephone | string | Yes | Telephone number in E.164 format |
| messages[].provider | string | Yes | UUID of the provider to use |
| messages[].refno | string | Yes | Reference number for tracking |
| messages[].categories | array | Yes | Array of category strings |
| messages[].identifiers | object | Yes | Identifiers for message tracking |
| messages[].identifiers.tenant | string | Yes | Tenant identifier |
| messages[].identifiers.eventUuid | string | No | Event UUID |
| messages[].identifiers.actionUuid | string | No | Action UUID |
| messages[].identifiers.actionCode | string | No | Action code |
| messages[].params | object | No | Template parameters |
| messages[].attachments | object | No | Message attachments |

**Response:**

```json
{
  "messages": [
    {
      "refno": "000000000001",
      "uuid": "a1b2c3d4-e5f6-7890-abcd-1234567890ab"
    }
  ]
}
```

## Template API

### `POST /api/v1/templates`

Create a new template.

**Request:**

```json
{
  "templates": [{
    "code": "alert-template",
    "name": "Alert Template",
    "content": "Hello {{name}}, there has been an alert in your area.",
    "status": 1,
    "channel": "WHATSAPP",
    "templateIds": {
      "twilio": "HM123456",
      "gupshup": "gupshup-template-1234"
    },
    "tenant": "example-tenant"
  }]
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| templates | array | Yes | Array containing at least one template object |
| templates[].code | string | Yes | Unique code for the template (unique per tenant, cannot be edited after creation) |
| templates[].name | string | Yes | Name of the template |
| templates[].content | string | Yes | Content of the template |
| templates[].status | number | No | Status of the template (0=inactive, 1=active). Default is 0 |
| templates[].channel | string | Yes | Channel for the template (WHATSAPP, SMS, EMAIL) |
| templates[].templateIds | object | No | Provider-specific template IDs as key-value pairs |
| templates[].tenant | string | Yes | Tenant identifier |

**Response:**

```json
{
  "message": "Templates created successfully",
  "templates": [
    {
      "uuid": "a1b2c3d4-e5f6-7890-abcd-1234567890ab",
      "code": "alert-template",
      "name": "Alert Template",
      "content": "Hello {{name}}, there has been an alert in your area.",
      "status": 1,
      "channel": "WHATSAPP",
      "templateIds": {
        "twilio": "HM123456",
        "gupshup": "gupshup-template-1234"
      },
      "tenant": "example-tenant",
      "createdAt": "2025-10-06T12:00:00Z",
      "updatedAt": "2025-10-06T12:00:00Z"
    }
  ]
}
```

### `GET /api/v1/templates`

Get all templates with optional filtering.

**Query Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| limit | integer | No | Maximum number of templates to return (default: 50) |
| offset | integer | No | Number of templates to skip for pagination |
| channel | string | No | Filter templates by channel (WHATSAPP, SMS, EMAIL) |
| tenant | string | No | Filter templates by tenant identifier |
| code | string | No | Filter templates by code |

**Response:**

```json
{
  "message": "Templates retrieved successfully",
  "templates": [
    {
      "uuid": "a1b2c3d4-e5f6-7890-abcd-1234567890ab",
      "name": "Alert Template",
      "content": "Hello {{name}}, there has been an alert in your area.",
      "status": 1,
      "channel": "WHATSAPP",
      "templateIds": {
        "twilio": "HM123456",
        "gupshup": "gupshup-template-1234"
      },
      "tenant": "example-tenant",
      "createdAt": "2025-10-06T12:00:00Z",
      "updatedAt": "2025-10-06T12:00:00Z"
    }
  ]
}
```

### `GET /api/v1/templates/{uuid}`

Get a template by UUID.

**Response:**

```json
{
  "message": "Template retrieved successfully",
  "templates": [
    {
      "uuid": "a1b2c3d4-e5f6-7890-abcd-1234567890ab",
      "name": "Alert Template",
      "content": "Hello {{name}}, there has been an alert in your area.",
      "status": 1,
      "channel": "WHATSAPP",
      "templateIds": {
        "twilio": "HM123456",
        "gupshup": "gupshup-template-1234"
      },
      "tenant": "example-tenant",
      "createdAt": "2025-10-06T12:00:00Z",
      "updatedAt": "2025-10-06T12:00:00Z"
    }
  ]
}
```

### `PUT /api/v1/templates/{uuid}`

Update a template by UUID.

**Request:**

```json
{
  "templates": [{
    "name": "Updated Alert Template",
    "content": "Hello {{name}}, there has been an important alert in your area.",
    "templateIds": {
      "twilio": "HM123456_UPDATED",
      "messagebird": "mb-template-5678"
    },
    "status": 1
    /* Note: code field cannot be edited after creation */
  }]
}
```

**Notes:**
- Only include fields that need to be updated
- To update the `templateIds` object, provide the complete object with all provider IDs you want to keep

**Response:**

```json
{
  "message": "Template updated successfully",
  "templates": [
    {
      "uuid": "a1b2c3d4-e5f6-7890-abcd-1234567890ab",
      "code": "alert-template",
      "name": "Updated Alert Template",
      "content": "Hello {{name}}, there has been an important alert in your area.",
      "status": 1,
      "channel": "WHATSAPP",
      "templateIds": {
        "twilio": "HM123456_UPDATED",
        "messagebird": "mb-template-5678"
      },
      "tenant": "example-tenant",
      "createdAt": "2025-10-06T12:00:00Z",
      "updatedAt": "2025-10-06T13:15:00Z"
    }
  ]
}
```

## Provider API

### `POST /api/v1/providers`

Create a new provider.

**Request:**

```json
{
  "providers": [{ 
    "code": "whatsapp-primary",
    "provider": "TWILIO",
    "name": "My Twilio WhatsApp Provider",
    "config": {
      "fromNumber": "+14155238886",
      "baseUrl": "https://api.twilio.com/2010-04-01",
      "accountSid": "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    },
    "secureConfig": {
      "authToken": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    },
    "channel": "WHATSAPP",
    "tenant": "default",
    "status": 1
  }]
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| providers | array | Yes | Array containing at least one provider object |
| providers[].code | string | Yes | Provider code (must be unique per tenant) |
| providers[].provider | string | Yes | Provider implementation class name (e.g., TWILIO, SENDGRID) |
| providers[].name | string | Yes | Provider name |
| providers[].config | object | Yes | Provider configuration (fields depend on provider type) |
| providers[].secureConfig | object | Yes | Secure provider configuration (will be encrypted) |
| providers[].status | number | No | Status of the provider (0=inactive, 1=active). Default is 0 |
| providers[].channel | string | Yes | Channel for the provider (WHATSAPP, SMS, EMAIL) |
| providers[].tenant | string | Yes | Tenant identifier |

**Response:**

```json
{
  "code": 0,
  "message": "Providers created successfully",
  "providers": [
    {
      "uuid": "0bca5714-bceb-49a4-a4eb-e3afcec26328",
      "code": "whatsapp-primary",
      "provider": "TWILIO",
      "name": "My Twilio WhatsApp Provider",
      "config": {
        "fromNumber": "+14155238886",
        "baseUrl": "https://api.twilio.com/2010-04-01",
        "accountSid": "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
      },
      "channel": "WHATSAPP",
      "tenant": "default",
      "status": 1,
      "createdAt": "2025-10-06T12:00:00Z",
      "updatedAt": "2025-10-06T12:00:00Z"
    }
  ]
}
```

### `GET /api/v1/providers`

Get all providers.

**Response:**

```json
{
  "code": 0,
  "message": "Providers retrieved successfully",
  "providers": [
    {
      "uuid": "0bca5714-bceb-49a4-a4eb-e3afcec26328",
      "code": "whatsapp-primary",
      "provider": "TWILIO",
      "name": "Twilio WhatsApp Provider",
      "config": {
        "fromNumber": "+14155238886",
        "baseUrl": "https://api.twilio.com/2010-04-01",
        "accountSid": "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
      },
      "channel": "WHATSAPP",
      "tenant": "default",
      "status": 1,
      "createdAt": "2025-10-06T12:00:00Z",
      "updatedAt": "2025-10-06T12:00:00Z"
    }
  ]
}
```

### `GET /api/v1/providers/{uuid}`

Get a provider by UUID.

**Response:**

```json
{
  "code": 0,
  "message": "Provider retrieved successfully",
  "providers": [
    {
      "uuid": "0bca5714-bceb-49a4-a4eb-e3afcec26328",
      "code": "whatsapp-primary",
      "provider": "TWILIO",
      "name": "Twilio WhatsApp Provider",
      "config": {
        "fromNumber": "+14155238886",
        "baseUrl": "https://api.twilio.com/2010-04-01",
        "accountSid": "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
      },
      "channel": "WHATSAPP",
      "tenant": "default",
      "status": 1,
      "createdAt": "2025-10-06T12:00:00Z",
      "updatedAt": "2025-10-06T12:00:00Z"
    }
  ]
}
```

### `PUT /api/v1/providers/{uuid}`

Update a provider by UUID.

**Request:**

```json
{
  "providers": [{ 
    "name": "Updated Twilio Provider",
    "config": {
      "fromNumber": "+14155238886",
      "baseUrl": "https://api.twilio.com/2010-04-01"
    },
    "secureConfig": {
      "authToken": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    },
    "status": 1
  }]
}
```

**Notes:**
- The `code` field cannot be updated (it must remain unique per tenant)
- The `provider` field cannot be updated after creation
- Only include fields that need to be updated

**Response:**

```json
{
  "code": 0,
  "message": "Provider updated successfully",
  "providers": [
    {
      "uuid": "0bca5714-bceb-49a4-a4eb-e3afcec26328",
      "code": "whatsapp-primary",
      "provider": "TWILIO",
      "name": "Updated Twilio Provider",
      "config": {
        "fromNumber": "+14155238886",
        "baseUrl": "https://api.twilio.com/2010-04-01",
        "accountSid": "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
      },
      "channel": "WHATSAPP",
      "tenant": "default",
      "status": 1,
      "createdAt": "2025-10-06T12:00:00Z",
      "updatedAt": "2025-10-06T13:15:00Z"
    }
  ]
}
```

## Email API

### `POST /api/v1/emails/`

Send an email to one or more recipients.

**Request:**

```json
{
  "to": ["recipient@example.com", "another@example.com"],
  "subject": "Your email subject",
  "body": "Your email content goes here...",
  "isHtml": true,
  "attachments": [
    {
      "filename": "document.pdf",
      "contentType": "application/pdf",
      "content": "base64EncodedContent..."
    }
  ]
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| to | array of strings | Yes | Email recipients |
| subject | string | Yes | Email subject line |
| body | string | Yes | Email body content |
| isHtml | boolean | Yes | Whether the body is HTML (true) or plain text (false) |
| attachments | array of objects | No | Optional file attachments |

**Attachment Object:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| filename | string | Yes | Name of the attachment file |
| contentType | string | Yes | MIME type of the attachment |
| content | string | Yes | Base64-encoded file content |

**Response (Success):**

```json
{
  "code": 200,
  "message": "Email sent successfully"
}
```

**Response (Error):**

```json
{
  "code": 400,
  "message": "Invalid request body"
}
```

## SMS API

### `POST /api/v1/sms/`

Send an SMS to one recipient or bulk messages to multiple recipients.

**Single SMS Request:**

```json
{
  "to": "+15551234567",
  "message": "Your SMS message here"
}
```

**Bulk SMS Request:**

```json
{
  "to": ["+15551234567", "+15557654321"],
  "message": "Your bulk SMS message here"
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| to | string or array of strings | Yes | Phone number(s) in E.164 format (with "+" prefix) |
| message | string | Yes | SMS message content |

**Response (Success - Single):**

```json
{
  "code": 200,
  "message": "SMS sent successfully"
}
```

**Response (Success - Bulk):**

```json
{
  "code": 200,
  "message": "Bulk SMS sent successfully"
}
```

**Response (Error):**

```json
{
  "code": 400,
  "message": "Invalid phone number format"
}
```

## WhatsApp API

### `POST /api/v1/whatsapp/`

Send a WhatsApp message (text, media, or template).

**Text Message Request:**

```json
{
  "type": "text",
  "to": "+15551234567",
  "message": "Your WhatsApp message here"
}
```

**Media Message Request:**

```json
{
  "type": "media",
  "to": "+15551234567",
  "caption": "Image caption",
  "mediaType": "image",
  "mediaUrl": "https://example.com/image.jpg"
}
```

**Template Message Request:**

```json
{
  "type": "template",
  "to": "+15551234567",
  "templateName": "appointment_reminder",
  "params": {
    "name": "John Doe",
    "date": "2025-10-10",
    "time": "14:30"
  }
}
```

**Parameters (Common):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| type | string | Yes | Message type: "text", "media", or "template" |
| to | string | Yes | Phone number in E.164 format (with "+" prefix) |

**Parameters (Text Message):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| message | string | Yes | Text message content |

**Parameters (Media Message):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| caption | string | No | Optional caption for the media |
| mediaType | string | Yes | Type of media: "image", "video", "document", etc. |
| mediaUrl | string | Yes | URL where the media is hosted |

**Parameters (Template Message):**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| templateName | string | Yes | Name of the pre-approved template |
| params | object | Yes | Key-value pairs for template variables |

**Response (Success):**

```json
{
  "code": 200,
  "message": "WhatsApp message sent successfully"
}
```

**Response (Error):**

```json
{
  "code": 400,
  "message": "Invalid message type"
}
```
