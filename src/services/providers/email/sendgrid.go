package email

import (
	"bytes"
	"delivery/api/types"
	"delivery/helper"
	"delivery/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SendGridProvider implements the EmailService interface using SendGrid
type SendGridProvider struct {
	APIKey    string
	FromEmail string
	BaseURL   string
	Client    *http.Client
	Provider  *models.Provider
}

// Config holds the SendGrid provider configuration
type Config struct {
	FromEmail string `json:"from,omitempty"`
	AccountID string `json:"accountId,omitempty"`
	BaseURL   string `json:"baseUrl,omitempty"`
}

// SecureConfig holds the SendGrid provider secure configuration
type SecureConfig struct {
	APIKey string `json:"apikey,omitempty"`
}

// NewSendGridProviderFromDB creates a new SendGrid email provider using database configuration
func NewSendGridProviderFromDB(provider *models.Provider) (*SendGridProvider, error) {
	if provider == nil {
		return nil, errors.New("provider cannot be nil")
	}

	// Parse config
	var config Config
	configJSON, err := json.Marshal(provider.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal provider config: %w", err)
	}

	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to parse provider config: %w", err)
	}

	// Parse secure config
	secureConfigJSON, err := json.Marshal(provider.SecureConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal provider secure config: %w", err)
	}

	// Get encryption key from environment
	encryptionKey := []byte(helper.GetEnv("ENCRYPTION_KEY", ""))
	if len(encryptionKey) != 32 {
		return nil, errors.New("ENCRYPTION_KEY environment variable not set or invalid (must be exactly 32 bytes)")
	}

	// Check if the config is encrypted
	var encryptedData struct {
		Encrypted string `json:"encrypted"`
	}
	if err := json.Unmarshal(secureConfigJSON, &encryptedData); err != nil || encryptedData.Encrypted == "" {
		return nil, fmt.Errorf("secure config is not in the expected format: %w", err)
	}

	// Decrypt the secure config
	decodedBytes, err := helper.DecodeBase64AndDecrypt(encryptedData.Encrypted, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt provider secure config: %w", err)
	}

	// Parse the decrypted secure config
	var secureConfig SecureConfig
	if err := json.Unmarshal(decodedBytes, &secureConfig); err != nil {
		return nil, fmt.Errorf("failed to parse provider secure config: %w", err)
	}

	// Validate required fields
	if secureConfig.APIKey == "" {
		return nil, errors.New("API key not set in provider configuration")
	}

	// Get the fromEmail
	fromEmail := config.FromEmail
	if fromEmail == "" {
		return nil, errors.New("from email not set in provider configuration")
	}

	// Get the baseURL
	baseURL := config.BaseURL
	if baseURL == "" {
		return nil, errors.New("base URL not set in provider configuration")
	}

	// Trim whitespace from API key
	apiKey := strings.TrimSpace(secureConfig.APIKey)

	return &SendGridProvider{
		APIKey:    apiKey,
		FromEmail: fromEmail,
		BaseURL:   baseURL,
		Client:    &http.Client{Timeout: 10 * time.Second},
		Provider:  provider,
	}, nil
}

// Send implements the EmailService.Send method
func (p *SendGridProvider) Send(to []string, subject string, body string, isHTML bool) error {
	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	emailRequest := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": toRecipientFormat(to),
			},
		},
		"from": map[string]string{
			"email": p.FromEmail,
		},
		"subject": subject,
		"content": []map[string]string{
			{
				"type":  contentType,
				"value": body,
			},
		},
	}

	return p.sendRequest(emailRequest)
}

// SendWithAttachments sends an email with attachments
func (p *SendGridProvider) SendWithAttachments(to []string, subject string, body string, isHTML bool, attachments []types.EmailAttachment) error {
	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	// Format the attachments for SendGrid API
	sendgridAttachments := []map[string]string{}
	for _, attachment := range attachments {
		encodedContent := helper.Base64Encode(attachment.Content)
		sendgridAttachments = append(sendgridAttachments, map[string]string{
			"content":     encodedContent,
			"type":        attachment.ContentType,
			"filename":    attachment.Filename,
			"disposition": "attachment",
		})
	}

	emailRequest := map[string]interface{}{
		"personalizations": []map[string]interface{}{
			{
				"to": toRecipientFormat(to),
			},
		},
		"from": map[string]string{
			"email": p.FromEmail,
		},
		"subject": subject,
		"content": []map[string]string{
			{
				"type":  contentType,
				"value": body,
			},
		},
		"attachments": sendgridAttachments,
	}

	return p.sendRequest(emailRequest)
}

// sendRequest sends a request to the SendGrid API
func (p *SendGridProvider) sendRequest(emailRequest map[string]interface{}) error {
	// Construct the endpoint from the BaseURL
	// Ensure the BaseURL doesn't end with a slash before appending the path
	endpoint := strings.TrimSuffix(p.BaseURL, "/") + "/v3/mail/send"

	requestBody, err := json.Marshal(emailRequest)
	if err != nil {
		return err
	}

	// Log detailed request info
	requestBodyStr := string(requestBody)
	// Mask API key in logs if it appears in the request body
	if p.APIKey != "" {
		requestBodyStr = strings.Replace(requestBodyStr, p.APIKey, "[REDACTED]", -1)
	}

	helper.Log.WithFields(map[string]interface{}{
		"endpoint":    endpoint,
		"fromEmail":   p.FromEmail,
		"recipients":  emailRequest["personalizations"].([]map[string]interface{})[0]["to"],
		"subject":     emailRequest["subject"],
		"requestBody": requestBodyStr,
		"contentType": emailRequest["content"].([]map[string]string)[0]["type"],
	}).Info("Sending SendGrid Email API request")

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read response body
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Create response fields for logging
	responseFields := map[string]interface{}{
		"statusCode": resp.StatusCode,
		"response":   bodyStr,
	}

	if resp.StatusCode >= 400 {
		// If we can parse the error as JSON, extract the errors
		var errorResponse struct {
			Errors []struct {
				Message string `json:"message"`
				Field   string `json:"field"`
				Help    string `json:"help"`
			} `json:"errors"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			for _, e := range errorResponse.Errors {
				responseFields[fmt.Sprintf("error_%s", e.Field)] = e.Message
			}
			helper.Log.WithFields(responseFields).Error("SendGrid API returned an error response")
		} else {
			helper.Log.WithFields(responseFields).Error("SendGrid API returned a non-JSON error response")
		}

		return fmt.Errorf("sendgrid API error: %s, status code: %d", bodyStr, resp.StatusCode)
	} else {
		// Log successful response
		helper.Log.WithFields(responseFields).Info("SendGrid API request successful")
	}

	return nil
}

// Helper function to convert string array of emails to SendGrid recipient format
func toRecipientFormat(emails []string) []map[string]string {
	recipients := make([]map[string]string, len(emails))
	for i, email := range emails {
		recipients[i] = map[string]string{
			"email": email,
		}
	}
	return recipients
}

// GetStatus gets the status of an email message
func (p *SendGridProvider) GetStatus(messageID string) (types.DeliveryStatus, error) {
	// SendGrid doesn't provide a direct way to get message status by ID
	return types.DeliveryStatus{
		MessageID: messageID,
		Status:    "unknown",
		Details:   "SendGrid provider does not support status retrieval by message ID",
		Timestamp: time.Now().Format(time.RFC3339),
	}, nil
}
