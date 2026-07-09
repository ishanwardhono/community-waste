package apperr

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

func New(code int, msg string) error {
	return &AppError{Code: code, Message: msg}
}
