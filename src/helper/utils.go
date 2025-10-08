package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
)

// Response is an alias to APIResponse for backward compatibility
type Response = APIResponse

// RespondWithJSON sends a JSON response (wrapper for backward compatibility)
func RespondWithJSON(w http.ResponseWriter, statusCode int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// GetEnv gets an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// RenderTemplate renders a Go template string with provided variables
// It takes a template content string and a map of variables to inject
// Returns the rendered string or an error
func RenderTemplate(templateContent string, variables map[string]string) (string, error) {
	// Create a custom function to get variables by key name
	// This handles all variable names, including numeric ones
	funcMap := template.FuncMap{
		"var": func(key string) string {
			if val, ok := variables[key]; ok {
				return val
			}
			return ""
		},
	}

	// Create a modified template that uses the var function to access variables
	// Replace {{key}} with {{var "key"}} in the template content
	modifiedTemplate := templateContent
	for key := range variables {
		placeholder := "{{" + key + "}}"
		replacement := "{{var \"" + key + "\"}}"
		modifiedTemplate = strings.Replace(modifiedTemplate, placeholder, replacement, -1)
	}

	// Parse the modified template with our custom var function
	tmpl, err := template.New("message").Funcs(funcMap).Parse(modifiedTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template with an empty data map since we're using the var function
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return buf.String(), nil
}
