package helper

import (
	"bytes"
	"encoding/base64"
	"html/template"
	"time"
)

// ProcessTemplate processes a template string with the given parameters
func ProcessTemplate(templateString string, params map[string]string) (string, error) {
	// Create a custom function map to support the {{var "key"}} syntax
	funcMap := template.FuncMap{
		"var": func(key string) string {
			if val, ok := params[key]; ok {
				return val
			}
			return ""
		},
	}

	// Get key names from the map, prioritizing string keys over numeric keys
	// This is helpful for templates that use simple {{.name}} syntax
	mappedParams := make(map[string]interface{})
	for k, v := range params {
		mappedParams[k] = v
	}

	// Create the template with our functions
	tmpl, err := template.New("content").Funcs(funcMap).Parse(templateString)
	if err != nil {
		return "", err
	}

	// First try executing with mappedParams to support {{.key}} syntax
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, mappedParams)

	// If there's an error, log it for debugging
	if err != nil {
		// Try executing with nil to support {{var "key"}} syntax
		buf.Reset()
		err = tmpl.Execute(&buf, nil)
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// DecodeBase64 decodes a base64 encoded string to bytes
func DecodeBase64(encoded string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(encoded)
}

// GetCurrentTime returns the current time
func GetCurrentTime() time.Time {
	return time.Now().UTC()
}

// DebugTemplate attempts to process a template and returns detailed information about any failures
func DebugTemplate(templateString string, params map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	result["template"] = templateString
	result["params"] = params

	// Try using standard template processing
	tmpl, err := template.New("content").Parse(templateString)
	if err != nil {
		result["parseError"] = err.Error()
		return result
	}

	// Try executing with params
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	if err != nil {
		result["executeError"] = err.Error()
	} else {
		result["standardResult"] = buf.String()
	}

	// Try with var function
	funcMap := template.FuncMap{
		"var": func(key string) string {
			if val, ok := params[key]; ok {
				return val
			}
			return ""
		},
	}

	tmpl, err = template.New("content").Funcs(funcMap).Parse(templateString)
	if err != nil {
		result["varFuncParseError"] = err.Error()
		return result
	}

	buf.Reset()
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		result["varFuncExecuteError"] = err.Error()
	} else {
		result["varFuncResult"] = buf.String()
	}

	return result
}
