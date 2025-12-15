package utils

import (
	"encoding/json"
	"net/http"
)

// ----------------------
// RESPONSE STRUCTS
// ----------------------

type SuccessResponse struct {
	Success bool   `json:"success"`           // TRUE
	Message string `json:"message,omitempty"` // Optional
	Count   int    `json:"count,omitempty"`   // Optional
	Data    any    `json:"data,omitempty"`    // Optional
	Meta    any    `json:"meta,omitempty"`    // Optional (pagination etc.)
}

type ErrorResponse struct {
	Success bool   `json:"success"` // FALSE
	Message string `json:"message"`
	Error   any    `json:"error,omitempty"`
}

// ----------------------
// BASE JSON WRITER
// ----------------------

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// ----------------------
// SUCCESS RESPONSES
// ----------------------

func Success(w http.ResponseWriter, message string, data any) {
	WriteJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithCount(w http.ResponseWriter, message string, count int, data any) {
	WriteJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
		Count:   count,
		Data:    data,
	})
}

func SuccessWithMeta(w http.ResponseWriter, message string, data any, meta any) {
	WriteJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// ----------------------
// ERROR RESPONSES (ALWAYS 200)
// ----------------------

func AppError(w http.ResponseWriter, msg string, err error) {
	WriteJSON(w, http.StatusOK, ErrorResponse{
		Success: false,
		Message: msg,
		Error:   errorToString(err),
	})
}

func Error (w http.ResponseWriter, msg string, err error) {
	AppError(w, msg, err)
}

func Http400(w http.ResponseWriter, err error, msg string) {
	AppError(w, msg, err)
}

func Http404(w http.ResponseWriter, msg string) {
	AppError(w, msg, nil)
}

// ----------------------
// REAL SERVER ERROR (500)
// ----------------------

func Http500(w http.ResponseWriter, err error) {
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
		Success: false,
		Message: "Internal server error",
		Error:   errorToString(err),
	})
}

// ----------------------
// HELPER
// ----------------------

func errorToString(err error) any {
	if err == nil {
		return nil
	}
	return err.Error()
}
