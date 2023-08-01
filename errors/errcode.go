package errors

const (
	// BadRequest 通用错误码，客户端请求错误
	BadRequestErr = 400000000
	// InternalError 通用错误码，服务端内部错误
	InternalErr = 500000000
	// Unauthorized 通用错误码，未授权或者授权已过期
	UnauthorizedErr = 401000000
	// ResourceNotFoundErr 通用错误码，请求资源不存在
	ResourceNotFoundErr = 404000000
	// Forbidden 通用错误码，禁止访问
	ForbiddenErr = 403000000
	// Conflict 通用错误码，资源冲突
	ConflictErr = 409000000
)