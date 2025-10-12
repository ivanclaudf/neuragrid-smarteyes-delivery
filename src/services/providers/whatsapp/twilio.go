package whatsapp

import (
	"bytes"
	apitypes "delivery/api/types"
	"delivery/helper"
	"delivery/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TwilioProvider implements the WhatsAppService interface using Twilio
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

// NewTwilioProviderFromDB creates a new Twilio WhatsApp provider using database configuration
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

	// Ensure fromNumber has whatsapp: prefix
	if !strings.HasPrefix(fromNumber, "whatsapp:") {
		fromNumber = "whatsapp:" + fromNumber
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

// SendText implements the WhatsAppService.SendText method
func (p *TwilioProvider) SendText(to string, message string) error {
	// Ensure to has whatsapp: prefix
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}

	formData := url.Values{}
	formData.Set("From", p.FromNumber)
	formData.Set("To", to)
	formData.Set("Body", message)

	return p.sendRequest(formData)
}

// SendMedia implements the WhatsAppService.SendMedia method
func (p *TwilioProvider) SendMedia(to string, caption string, mediaType string, mediaURL string) error {
	// Ensure to has whatsapp: prefix
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}

	formData := url.Values{}
	formData.Set("From", p.FromNumber)
	formData.Set("To", to)
	if caption != "" {
		formData.Set("Body", caption)
	}
	formData.Set("MediaUrl", mediaURL)

	return p.sendRequest(formData)
}

// SendTemplate implements the WhatsAppService.SendTemplate method
func (p *TwilioProvider) SendTemplate(to string, templateName string, params map[string]string) error {
	// Ensure to has whatsapp: prefix
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}

	// Use the provided template ID - this should be the provider-specific template ID
	// that has already been mapped in the consumer
	contentSid := templateName

	helper.Log.WithField("twilioTemplateID", contentSid).Debug("Using template ID for Twilio message")

	// Convert params to a JSON string for Twilio's template variables
	var contentVariables string
	if len(params) > 0 {
		paramsJSON, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal template parameters: %w", err)
		}
		contentVariables = string(paramsJSON)
	} else {
		contentVariables = "{}"
	}

	// Set up the form data for the API request
	formData := url.Values{}
	formData.Set("From", p.FromNumber)
	formData.Set("To", to)
	formData.Set("ContentSid", contentSid)
	formData.Set("ContentVariables", contentVariables)

	// Note: We're not setting the Body field here because the content is passed
	// through the ContentSid and ContentVariables fields for Twilio's WhatsApp templates.
	// The template rendering happens at the consumer level before reaching this function.

	return p.sendRequest(formData)
}

// sendRequest sends a request to the Twilio API
func (p *TwilioProvider) sendRequest(formData url.Values) error {
	endpoint := fmt.Sprintf("%s/Accounts/%s/Messages.json", p.BaseURL, p.AccountSID)

	// Log basic request info
	helper.Log.WithFields(map[string]interface{}{
		"endpoint":   endpoint,
		"from":       formData.Get("From"),
		"to":         formData.Get("To"),
		"contentSid": formData.Get("ContentSid"),
	}).Debug("Sending Twilio API request")

	// Validate contentSid format for Twilio templates
	if contentSid := formData.Get("ContentSid"); contentSid != "" {
		if !strings.HasPrefix(contentSid, "HX") {
			helper.Log.WithField("contentSid", contentSid).Warn("ContentSid may not be in the correct format for Twilio WhatsApp templates (should typically start with 'HX')")
		}
	}

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
			"contentSid": formData.Get("ContentSid"),
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

			if errorResponse.Code == 20422 {
				return fmt.Errorf("twilio API invalid parameter error (code 20422) - likely an invalid template ID. Check that the template ID is correctly configured for provider '%s' and is a valid Twilio template ID. Response: %s", p.Provider.Provider, bodyStr)
			}
		} else {
			helper.Log.WithFields(errFields).Error("Twilio API returned a non-JSON error response")
		}

		return fmt.Errorf("twilio API error: %s, status code: %d", bodyStr, resp.StatusCode)
	}

	return nil
}

// GetStatus implements the WhatsAppService.GetStatus method
func (p *TwilioProvider) GetStatus(messageID string) (apitypes.DeliveryStatus, error) {
	endpoint := fmt.Sprintf("%s/Accounts/%s/Messages/%s.json", p.BaseURL, p.AccountSID, messageID)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return apitypes.DeliveryStatus{}, err
	}

	req.SetBasicAuth(p.AccountSID, p.AuthToken)

	resp, err := p.Client.Do(req)
	if err != nil {
		return apitypes.DeliveryStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return apitypes.DeliveryStatus{}, fmt.Errorf("twilio API error: %s, status code: %d", string(body), resp.StatusCode)
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
		return apitypes.DeliveryStatus{}, err
	}

	if err := json.Unmarshal(buf.Bytes(), &messageData); err != nil {
		return apitypes.DeliveryStatus{}, err
	}

	details := ""
	if messageData.ErrorCode != "" {
		details = fmt.Sprintf("Error code: %s, Error message: %s", messageData.ErrorCode, messageData.ErrorMessage)
	}

	return apitypes.DeliveryStatus{
		MessageID: messageID,
		Status:    messageData.Status,
		Details:   details,
		Timestamp: messageData.DateUpdated,
	}, nil
}
