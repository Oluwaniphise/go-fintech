package common

type Response[T any] struct {
	Success    bool   `json:"success"`
	StatusCode int    `json:"statusCode"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Data       T      `json:"data"`
	Error      any    `json:"error,omitempty"`
}

type ErrorDetail struct {
	Details string `json:"details"`
}

func Success[T any](statusCode int, code, message string, data T) Response[T] {
	return Response[T]{
		Success:    true,
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Data:       data,
	}
}

func Failure(statusCode int, code, message string, err any) Response[any] {
	return Response[any]{
		Success:    false,
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Error:      err,
	}
}
