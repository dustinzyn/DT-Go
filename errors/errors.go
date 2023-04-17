package errors

import (
	"net/http"

	"devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/GoCommon/api"
)
const publicServiceCode int = 0

var (
	serviceErrBuilder    api.TypedErrorBuilder
	InternalServerError  func(*api.ErrorInfo) *api.Error
	BadRequestError      func(*api.ErrorInfo) *api.Error
	UnauthorizationError func(*api.ErrorInfo) *api.Error
	NoPermissionError    func(*api.ErrorInfo) *api.Error
	NotFoundError        func(*api.ErrorInfo) *api.Error
)

func init() {
	serviceErrBuilder = api.NewErrorBuilder(publicServiceCode)

	// 500
	InternalServerError = serviceErrBuilder.OfType(&api.ErrorType{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  0,
		ErrorInfo: api.ErrorInfo{
			Message: "Internal server error.",
		},
	})
	// 400
	BadRequestError = serviceErrBuilder.OfType(&api.ErrorType{
		StatusCode: http.StatusBadRequest,
		ErrorCode:  0,
		ErrorInfo: api.ErrorInfo{
			Message: "Invalid request.",
		},
	})
	// 401
	UnauthorizationError = serviceErrBuilder.OfType(&api.ErrorType{
		StatusCode: http.StatusUnauthorized,
		ErrorCode:  0,
		ErrorInfo: api.ErrorInfo{
			Message: "Not authorized.",
		},
	})
	// 403
	NoPermissionError = serviceErrBuilder.OfType(&api.ErrorType{
		StatusCode: http.StatusForbidden,
		ErrorCode:  0,
		ErrorInfo: api.ErrorInfo{
			Message: "No permission to do this service.",
		},
	})
	// 404
	NotFoundError = serviceErrBuilder.OfType(&api.ErrorType{
		StatusCode: http.StatusNotFound,
		ErrorCode:  0,
		ErrorInfo: api.ErrorInfo{
			Message: "Resource not found.",
		},
	})
}
