package sms

import (
	"bytes"
	"delivery/helper"
	"delivery/models"
	"delivery/services"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TwilioProvider implements the SMSService interface using Twilio
type TwilioProvider struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	BaseURL    string
	Client     *http.Client
	Provider   *models.Provider
}

// Config holds the Twilio provider configuration
type Config struct {
	BaseURL    string `json:"baseUrl,omitempty"`
	FromNumber string `json:"fromNumber,omitempty"`
	AccountSID string `json:"accountSid,omitempty"`
}

// SecureConfig holds the Twilio provider secure configuration
type SecureConfig struct {
	AuthToken string `json:"authToken,omitempty"`
}

// NewTwilioProvider creates a new Twilio SMS provider from environment variables
func NewTwilioProvider() (*TwilioProvider, error) {
	accountSID := helper.GetEnv("TWILIO_ACCOUNT_SID", "")
	authToken := helper.GetEnv("TWILIO_AUTH_TOKEN", "")
	fromNumber := helper.GetEnv("TWILIO_FROM_NUMBER", "")
	baseURL := helper.GetEnv("TWILIO_BASE_URL", "https://api.twilio.com/2010-04-01")

	if accountSID == "" {
		return nil, errors.New("TWILIO_ACCOUNT_SID environment variable not set")
	}

	if authToken == "" {
		return nil, errors.New("TWILIO_AUTH_TOKEN environment variable not set")
	}

	if fromNumber == "" {
		return nil, errors.New("TWILIO_FROM_NUMBER environment variable not set")
	}

	return &TwilioProvider{
		AccountSID: accountSID,
		AuthToken:  authToken,
		FromNumber: fromNumber,
		BaseURL:    baseURL,
		Client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// NewTwilioProviderFromDB creates a new Twilio SMS provider using database configuration
func NewTwilioProviderFromDB(provider *models.Provider) (*TwilioProvider, error) {
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
		// For development, use a fixed key if not provided
		encryptionKey = []byte("12345678901234567890123456789012")
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
	if config.AccountSID == "" {
		return nil, errors.New("account SID not set in provider configuration")
	}

	if secureConfig.AuthToken == "" {
		return nil, errors.New("auth token not set in provider configuration")
	}

	// Check for placeholder auth tokens
	placeholders := []string{"your-auth-token-here", "your_auth_token_here", "your_auth_token", "your-auth-token", "auth_token_here"}
	for _, placeholder := range placeholders {
		if strings.Contains(secureConfig.AuthToken, placeholder) {
			return nil, errors.New("auth token contains placeholder value, please update with a real Twilio auth token")
		}
	}

	// Get the fromNumber
	fromNumber := config.FromNumber
	if fromNumber == "" {
		return nil, errors.New("from number not set in provider configuration")
	}

	// Get the baseURL
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.twilio.com/2010-04-01"
	}

	// Trim whitespace from auth token
	authToken := strings.TrimSpace(secureConfig.AuthToken)

	return &TwilioProvider{
		AccountSID: config.AccountSID,
		AuthToken:  authToken,
		FromNumber: fromNumber,
		BaseURL:    baseURL,
		Client:     &http.Client{Timeout: 10 * time.Second},
		Provider:   provider,
	}, nil
}

// Send implements the SMSService.Send method
func (p *TwilioProvider) Send(to string, message string) error {
	formData := url.Values{}
	formData.Set("From", p.FromNumber)
	formData.Set("To", to)
	formData.Set("Body", message)

	return p.sendRequest(formData)
}

// SendBulk implements the SMSService.SendBulk method
func (p *TwilioProvider) SendBulk(to []string, message string) error {
	// Send SMS to each recipient
	var lastErr error
	for _, recipient := range to {
		if err := p.Send(recipient, message); err != nil {
			lastErr = err
			helper.Log.WithError(err).WithField("recipient", recipient).Error("Failed to send SMS to recipient")
		}
	}
	return lastErr
}

// SendTemplate implements the SMSService.SendTemplate method
// For now, we just use the rendered content from the template (done earlier)
// and send it via the normal Send method. In the future, this could use
// provider-specific template APIs if available.
func (p *TwilioProvider) SendTemplate(to string, templateName string, params map[string]string) error {
	// If the rendered content was provided in the params, use it
	renderedContent, exists := params["rendered_content"]
	if !exists {
		return errors.New("rendered_content not found in params")
	}

	helper.Log.WithFields(map[string]interface{}{
		"to":           to,
		"template":     templateName,
		"content":      renderedContent,
		"provider":     "twilio",
		"provider_sid": p.AccountSID,
	}).Debug("Sending template SMS via Twilio")

	return p.Send(to, renderedContent)
}

// sendRequest sends a request to the Twilio API
func (p *TwilioProvider) sendRequest(formData url.Values) error {
	endpoint := fmt.Sprintf("%s/Accounts/%s/Messages.json", p.BaseURL, p.AccountSID)

	// Log basic request info
	helper.Log.WithFields(map[string]interface{}{
		"endpoint": endpoint,
		"from":     formData.Get("From"),
		"to":       formData.Get("To"),
	}).Debug("Sending Twilio SMS API request")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(p.AccountSID, p.AuthToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		// Log more detailed error information
		errFields := map[string]interface{}{
			"statusCode": resp.StatusCode,
			"response":   bodyStr,
		}

		// If we can parse the error as JSON, extract the code and message
		var errorResponse struct {
			Code     int    `json:"code"`
			Message  string `json:"message"`
			MoreInfo string `json:"more_info"`
		}

		if err := json.Unmarshal(body, &errorResponse); err == nil {
			errFields["errorCode"] = errorResponse.Code
			errFields["errorMessage"] = errorResponse.Message
			errFields["moreInfo"] = errorResponse.MoreInfo

			helper.Log.WithFields(errFields).Error("Twilio API returned an error response")
		} else {
			helper.Log.WithFields(errFields).Error("Twilio API returned a non-JSON error response")
		}

		return fmt.Errorf("twilio API error: %s, status code: %d", bodyStr, resp.StatusCode)
	}

	return nil
}

// GetStatus implements the SMSService.GetStatus method
func (p *TwilioProvider) GetStatus(messageID string) (services.DeliveryStatus, error) {
	endpoint := fmt.Sprintf("%s/Accounts/%s/Messages/%s.json", p.BaseURL, p.AccountSID, messageID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return services.DeliveryStatus{}, err
	}

	req.SetBasicAuth(p.AccountSID, p.AuthToken)

	resp, err := p.Client.Do(req)
	if err != nil {
		return services.DeliveryStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return services.DeliveryStatus{}, fmt.Errorf("twilio API error: %s, status code: %d", string(body), resp.StatusCode)
	}

	var messageData struct {
		Status       string `json:"status"`
		ErrorCode    string `json:"error_code"`
		ErrorMessage string `json:"error_message"`
		DateUpdated  string `json:"date_updated"`
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return services.DeliveryStatus{}, err
	}

	if err := json.Unmarshal(buf.Bytes(), &messageData); err != nil {
		return services.DeliveryStatus{}, err
	}

	details := ""
	if messageData.ErrorCode != "" {
		details = fmt.Sprintf("Error code: %s, Error message: %s", messageData.ErrorCode, messageData.ErrorMessage)
	}

	return services.DeliveryStatus{
		MessageID: messageID,
		Status:    messageData.Status,
		Details:   details,
		Timestamp: messageData.DateUpdated,
	}, nil
}
