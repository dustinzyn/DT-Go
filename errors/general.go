// 通用错误码 500 400 401 403 404
package errors

import (
	"net/http"
)

const publicServiceCode int = 0

var (
	serviceErrBuilder    TypedErrorBuilder
	InternalServerError  func(*ErrorInfo) *Error
	BadRequestError      func(*ErrorInfo) *Error
	UnauthorizationError func(*ErrorInfo) *Error
	NoPermissionError    func(*ErrorInfo) *Error
	NotFoundError        func(*ErrorInfo) *Error
)

func init() {
	serviceErrBuilder = NewErrorBuilder(publicServiceCode)

	// 500
	InternalServerError = serviceErrBuilder.OfType(&ErrorType{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  0,
		ErrorInfo: ErrorInfo{
			Message:     "Internal server error.",
			Description: "Internal server error.",
		},
	})
	// 400
	BadRequestError = serviceErrBuilder.OfType(&ErrorType{
		StatusCode: http.StatusBadRequest,
		ErrorCode:  0,
		ErrorInfo: ErrorInfo{
			Message:     "Invalid request.",
			Description: "Invalid request.",
		},
	})
	// 401
	UnauthorizationError = serviceErrBuilder.OfType(&ErrorType{
		StatusCode: http.StatusUnauthorized,
		ErrorCode:  0,
		ErrorInfo: ErrorInfo{
			Message:     "Not authorized.",
			Description: "Not authorized.",
		},
	})
	// 403
	NoPermissionError = serviceErrBuilder.OfType(&ErrorType{
		StatusCode: http.StatusForbidden,
		ErrorCode:  0,
		ErrorInfo: ErrorInfo{
			Message:     "No permission to do this service.",
			Description: "No permission to do this service.",
		},
	})
	// 404
	NotFoundError = serviceErrBuilder.OfType(&ErrorType{
		StatusCode: http.StatusNotFound,
		ErrorCode:  0,
		ErrorInfo: ErrorInfo{
			Message:     "Resource not found.",
			Description: "Resource not found.",
		},
	})
}
