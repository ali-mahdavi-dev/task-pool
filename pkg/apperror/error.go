package apperror

func NotFound(message string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: message,
		Details: "",
	}
}

func BadRequest(message string) *AppError {
	return &AppError{
		Code:    "BAD_REQUEST",
		Message: message,
		Details: "",
	}
}

func InternalServerError(message string) *AppError {
	return &AppError{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: message,
		Details: "",
	}
}
