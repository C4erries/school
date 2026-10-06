package response

import (
	"github.com/labstack/echo/v4"

	"github.com/C4erries/school/backend/internal/infrastructure/api/adapters/http/generated"
)

// Error отправляет клиенту стандартизированный JSON-ответ с ошибкой.
func Error(c echo.Context, status int, code, message string) error {
	resp := generated.ErrorResponse{
		Error: generated.ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	return c.JSON(status, resp)
}

