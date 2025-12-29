package apperror

import (
	"github.com/gofiber/fiber/v2"
)

func HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	appErr, ok := err.(*AppError)
	if !ok {
		appErr = &AppError{
			Code:    "INTERNAL_SERVER_ERROR",
			Status:  500,
			Message: "Internal server error",
			Details: err.Error(),
		}
	}

	return c.Status(appErr.Status).JSON(fiber.Map{
		"code":    appErr.Code,
		"message": appErr.Message,
		"details": appErr.Details,
	})
}
