package utils

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
	Meta   *Meta       `json:"meta,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Status string       `json:"status"`
	Error  ErrorDetails `json:"error"`
}

// ErrorDetails contains error information
type ErrorDetails struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Meta contains metadata for responses (pagination, etc.)
type Meta struct {
	Page       int `json:"page,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
	Total      int `json:"total,omitempty"`
	Limit      int `json:"limit,omitempty"`
}

// RespondSuccess sends a successful JSON response
func RespondSuccess(w http.ResponseWriter, statusCode int, data interface{}, meta *Meta) {
	response := SuccessResponse{
		Status: "success",
		Data:   data,
		Meta:   meta,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// RespondError sends an error JSON response
func RespondError(w http.ResponseWriter, statusCode int, errorCode, message string, details interface{}) {
	response := ErrorResponse{
		Status: "error",
		Error: ErrorDetails{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// RespondValidationError sends a validation error response
func RespondValidationError(w http.ResponseWriter, fields map[string]string) {
	RespondError(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", fields)
}

// Helper functions for common errors
func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, "BAD_REQUEST", message, nil)
}

func RespondUnauthorized(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func RespondForbidden(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func RespondNotFound(w http.ResponseWriter, resource string) {
	message := resource + " not found"
	RespondError(w, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func RespondInternalError(w http.ResponseWriter) {
	RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred", nil)
}
