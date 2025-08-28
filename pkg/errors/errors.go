package errors

import (
	"errors"
	"fmt"
)

const (
	InternalServerError = "internal server error"
	BadRequest          = "bad request"
	NotFound            = "not_found"
	UnauthorizedError   = "unauthorized"

	UnauthorizedErrorCode   = 401
	InvalidDataCode         = 402
	ForbiddenErrorCode      = 403
	InternalServerErrorCode = 500
	NotFoundErrorCode       = 404
)

// AppError представляет собой стандартизированную структуру ошибки для API.
type AppError struct {
	Code         int    `json:"code"`    // HTTP статус код
	Message      string `json:"message"` // Сообщение для клиента
	Err          error  `json:"-"`       // Внутренняя ошибка, не для клиента
	IsUserFacing bool   `json:"-"`       // Флаг, указывающий, можно ли показывать `Err`
}

func (a *AppError) Error() string {
	if a == nil {
		return ""
	}
	if a.Err != nil {
		return fmt.Sprintf("%s (code: %d): %v", a.Message, a.Code, a.Err)
	}
	return fmt.Sprintf("%s (code: %d)", a.Message, a.Code)
}

// NewAppError создает новый экземпляр AppError.
func NewAppError(httpCode int, message string, err error, isUserFacing bool) *AppError {
	return &AppError{
		Code:         httpCode,
		Message:      message,
		Err:          err,
		IsUserFacing: isUserFacing,
	}
}

var (
	ErrDataNotFound = errors.New("data not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInternal     = errors.New("internal error")
)
