package utils

import "NEMBUS/internal/repository"

// Standard codes
const (
	CodeOK       = 200
	CodeCreated  = 201
	CodeNotFound = 404
	CodeBadReq   = 400
	CodeError    = 500
)

// NewResponse creates a standard response object
func NewResponse(statusCode int, message string, data interface{}) *repository.Response {
	return &repository.Response{
		StatusCode: statusCode,
		Message:    message,
		Data:       data,
	}
}
