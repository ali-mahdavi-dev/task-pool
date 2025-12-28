package apperror

import (
	"fmt"
)

type AppError struct {
	Code    string `json:"code"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func NewAppError(code string, status int, message, details string) *AppError {
	return &AppError{Code: code, Status: status, Message: message, Details: details}
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code: %s, message: %s, details: %s", e.Code, e.Message, e.Details)
}

func (e *AppError) Wrap(err error) *AppError {
	return &AppError{Code: e.Code, Status: e.Status, Message: e.Message, Details: fmt.Sprintf("%s: %s", e.Details, err.Error())}
}
