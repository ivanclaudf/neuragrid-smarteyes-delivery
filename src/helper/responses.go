package helper

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondWithError sends an error response with the given status code and message
func RespondWithError(w http.ResponseWriter, httpStatus int, code int, message string) {
	response := APIResponse{
		Code:    code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(response)
}

// RespondWithSuccess sends a success response with data
func RespondWithSuccess(w http.ResponseWriter, httpStatus int, message string, data interface{}) {
	response := APIResponse{
		Code:    CodeSuccess,
		Message: message,
	}

	// Instead of always wrapping the data, just append it to the response
	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    CodeServerError,
			Message: "Failed to marshal response",
		})
		return
	}

	// Convert response to map
	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseJSON, &responseMap); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    CodeServerError,
			Message: "Failed to process response",
		})
		return
	}

	// Convert data to map
	dataJSON, err := json.Marshal(data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    CodeServerError,
			Message: "Failed to marshal data",
		})
		return
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(dataJSON, &dataMap); err == nil {
		// If data successfully converts to map, merge it with response
		for k, v := range dataMap {
			responseMap[k] = v
		}
	} else {
		// If data can't be converted to map, add it as is
		responseMap["data"] = data
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(responseMap)
}

// WriteJSON writes the given data as JSON to the response writer
func WriteJSON(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
}

// RespondWithSuccessNoDataWrapper sends a success response without wrapping the data in a "data" key
func RespondWithSuccessNoDataWrapper(w http.ResponseWriter, httpStatus int, message string, data interface{}) {
	// Create response including code and message fields but placing data fields at top level
	response := struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    interface{} `json:"-"` // This won't be marshaled
	}{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	}

	// Marshal the response to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    CodeServerError,
			Message: "Failed to marshal response",
		})
		return
	}

	// Convert response to map (this will have code and message)
	var responseMap map[string]interface{}
	if err := json.Unmarshal(responseJSON, &responseMap); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    CodeServerError,
			Message: "Failed to process response",
		})
		return
	}

	// Marshal the data
	dataJSON, err := json.Marshal(data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{
			Code:    CodeServerError,
			Message: "Failed to marshal data",
		})
		return
	}

	// Unmarshal data to map
	var dataMap map[string]interface{}
	if err := json.Unmarshal(dataJSON, &dataMap); err == nil {
		// Merge data fields into response map
		for k, v := range dataMap {
			responseMap[k] = v
		}
	} else {
		// If data is not a structure that can be converted to a map,
		// fallback to the standard approach
		responseMap["data"] = data
	}

	// Special handling for models.ProviderResponse, ensure we don't wrap with "data"
	if _, ok := responseMap["providers"]; ok {
		// We already have providers directly at the top level, so we can remove "data" if it exists
		delete(responseMap, "data")
	}

	// Set headers and write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(responseMap)
}
