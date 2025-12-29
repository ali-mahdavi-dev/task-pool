package apperror

func NotFound(message string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Status:  404,
		Message: message,
		Details: "",
	}
}

func BadRequest(message string) *AppError {
	return &AppError{
		Code:    "BAD_REQUEST",
		Status:  400,
		Message: message,
		Details: "",
	}
}

func InternalServerError(message string) *AppError {
	return &AppError{
		Code:    "INTERNAL_SERVER_ERROR",
		Status:  500,
		Message: message,
		Details: "",
	}
}
