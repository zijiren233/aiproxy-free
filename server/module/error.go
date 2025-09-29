package module

import "net/http"

type OpenAIErrorResponse struct {
	Error OpenAIError `json:"error"`
}

type OpenAIError struct {
	Code    any    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Param   string `json:"param,omitempty"`
}

func NewOpenAIError(errorType, message string, code any) *OpenAIErrorResponse {
	return &OpenAIErrorResponse{
		Error: OpenAIError{
			Type:    errorType,
			Message: message,
			Code:    code,
		},
	}
}

func NewOpenAIErrorWithParam(errorType, message, param string, code any) *OpenAIErrorResponse {
	return &OpenAIErrorResponse{
		Error: OpenAIError{
			Type:    errorType,
			Message: message,
			Code:    code,
			Param:   param,
		},
	}
}

func NewInternalServerError() *OpenAIErrorResponse {
	return NewOpenAIError(
		"internal_server_error",
		"Internal server error",
		http.StatusInternalServerError,
	)
}

func NewInvalidRequestError(message string) *OpenAIErrorResponse {
	return NewOpenAIError("invalid_request_error", message, http.StatusBadRequest)
}

func NewRateLimitError(message string) *OpenAIErrorResponse {
	return NewOpenAIError("rate_limit_exceeded", message, http.StatusTooManyRequests)
}

func NewAuthenticationError(message string) *OpenAIErrorResponse {
	return NewOpenAIError("invalid_api_key", message, http.StatusUnauthorized)
}

func NewBadGatewayError(message string) *OpenAIErrorResponse {
	return NewOpenAIError("upstream_error", message, http.StatusBadGateway)
}
