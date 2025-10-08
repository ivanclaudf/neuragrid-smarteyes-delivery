package helper

const (
	CodeSuccess = 0
	MsgSuccess  = "Operation processed successfully"

	CodeError = 500
	MsgError  = "An error occurred while processing your request. Please try again later."

	CodeServerError = 500
	MsgServerError  = "An internal server error occurred."

	CodeDuplicate = 409
	MsgDuplicate  = "Duplicate record found."

	CodeBadRequest        = 400
	MsgBadRequest         = "Invalid request parameters."
	MsgInvalidRequestBody = "Invalid request body."

	CodeNotFound = 404
	MsgNotFound  = "Resource not found."

	// Time format for API responses
	TimeFormat = "2006-01-02T15:04:05Z07:00"
)
