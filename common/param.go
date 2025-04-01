package common

import (
	"errors"
	"fmt"
)

const (
	SynapseHttpOK     int = 0
	HttpOk            int = 200
	HttpInternalError int = 500
)

var (
	// The error code indicating that the system is paused.
	ErrSystemPaused     = &ApiResponse{Code: -100000, Msg: "This service is temporarily unavailable. Please contact customer support."}
	ErrBadArgument      = NewApiError(100001, "Bad argument")
	ErrTimeout          = NewApiError(100002, "Timeout")
	ErrUnauthorized     = NewApiError(110001, "Unauthorized")
	ErrEndpointNotFound = NewApiError(110002, "Endpoint not found")
	ErrInferenceError   = NewApiError(110003, "Inference error")
	ErrNoReadyClient    = NewApiError(110004, "No ready client")

	ErrSaasPodIDExists   = NewApiError(110005, "Agent service with saasPodID already exists.")
	ErrRequestLimit      = NewApiError(110006, "Request is too frequent.")
	ErrSaasPodProcessing = NewApiError(110007, "Request is being processed, please try again later.")
	ErrSaasPodIDNotExist = NewApiError(110008, "Pod with saasPodID does not exist.") // exists 改为 exist
)

func ConvertSaaSErrCode(code int, msg string) *ApiError {
	return &ApiError{
		Code: ApiErrorCode(code),
		Msg:  msg,
	}
}

type ApiErrorCode int

type ApiError struct {
	Code         ApiErrorCode   `json:"code"`
	Msg          string         `json:"msg"`
	TemplateData map[string]any `json:"-"`
}

func NewApiError(code ApiErrorCode, msg string) *ApiError {
	return &ApiError{
		Code: code,
		Msg:  msg,
	}
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("code: %v, msg: %v", e.Code, e.Msg)
}

type Validatable interface {
	// Validate do the validation for user request
	Validate() error
}

func Validate(r any) error {
	if validatable, ok := r.(Validatable); ok {
		return validatable.Validate()
	}
	return nil
}

type ApiResponse struct {
	Code         int            `json:"code"`
	Msg          string         `json:"msg"`
	Data         any            `json:"data"`
	TemplateData map[string]any `json:"-"`
}

func Ok(data any) *ApiResponse {
	return &ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: data,
	}
}

func GuessError(err error, unknownErrHandler func(error)) (httpCode int, apiResp *ApiResponse) {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return HttpOk, &ApiResponse{
			Code:         int(apiErr.Code),
			Msg:          apiErr.Msg,
			TemplateData: apiErr.TemplateData,
		}
	}
	unknownErrHandler(err)
	return HttpInternalError, internalError()
}

func internalError() *ApiResponse {
	return &ApiResponse{
		Code: 500,
		Msg:  "Internal error",
	}
}
