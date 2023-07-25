package errors

import (
	"fmt"
	"strings"
)

// Error API异常错误结构，会作为最终返回给前端的JSON Body
type Error struct {
	Code        int         `json:"code"`             // 错误码，前三位等于状态码，中间三位为服务标识，后三位为错误标识
	Message     string      `json:"message"`          // 错误消息
	Cause       string      `json:"cause"`            // 错误原因，产生错误的具体原因
	Detail      interface{} `json:"detail,omitempty"` // 错误码拓展信息，补充说明错误信息
	Description string      `json:"description"`      // 错误描述，可以和message保持一致，需要符合国际化要求，客户端采用此字段做错误提示
	Solution    string      `json:"solution"`         // 操作提示，针对当前错误的操作提示，需要符合国际化要求
}

// Error 实现error接口
//
// 暂时没有处理Detail
func (err *Error) Error() string {
	errInfo := []string{}

	if err.Code != 0 {
		errInfo = append(errInfo, fmt.Sprintf("Code: %d", err.Code))
	}

	if err.Message != "" {
		errInfo = append(errInfo, fmt.Sprintf("Message: %s", err.Message))
	}

	if err.Cause != "" {
		errInfo = append(errInfo, fmt.Sprintf("Cause: %s", err.Cause))
	}

	return strings.Join(errInfo, ", ")
}

// ErrorType 错误生成器参数对象
type ErrorType struct {
	StatusCode int // HTTP状态码，从net/http包中获得
	ErrorCode  int // 错误标识，不要使用0开头，否则会转为8进制
	ErrorInfo      // 错误详情，如果一类错误会返回固定的信息，可以传入此结构体；在构造实际的错误对象时还可覆盖。
}

// ErrorInfo 错误详情，是一些描述信的文本信息
type ErrorInfo struct {
	Message     string      `json:"message"`          // 错误消息
	Cause       string      `json:"cause"`            // 错误原因，产生错误的具体原因
	Detail      interface{} `json:"detail,omitempty"` // 错误码拓展信息，补充说明错误信息
	Description string      `json:"description"`      // 错误描述，可以和message保持一致，需要符合国际化要求，客户端采用此字段做错误提示
	Solution    string      `json:"solution"`         // 操作提示，针对当前错误的操作提示，需要符合国际化要求
}

// TypedErrorBuilder 固定Code，不同信息的错误生成器
type TypedErrorBuilder struct {
	serviceCode int          // 服务唯一标识，不要传入0开头，否则会转为8进制；尽量使用service.go中定义的Code
	usedCode    map[int]bool // 已经使用过的code，避免重复定义
}

// ErrorOfType 返回特定类型的错误函数
type ErrorOfType func(info *ErrorInfo) *Error

// OfType 返回固定Code的错误生成器
//
// 如果Code发生了冲突，则表明既有的Code表达不合理，需要修改。故重复Code会引起panic。
func (builder *TypedErrorBuilder) OfType(errType *ErrorType) ErrorOfType {
	code := buildCode(builder.serviceCode, errType.StatusCode, errType.ErrorCode)

	// code 必须为全局唯一，不允许定义两个一样的code
	if _, exists := builder.usedCode[code]; exists {
		panic(fmt.Errorf(`Fail to execute (*TypedErrorBuilder).OfType(): code %d has been declared`, code))
	}

	builder.usedCode[code] = true

	return func(info *ErrorInfo) *Error {
		err := Error{
			Code:    code,
			Message: errType.Message,
			Cause:   errType.Cause,
			Detail:  errType.Detail,
			Description: errType.Description,
			Solution: errType.Solution,
		}

		if info == nil {
			info = &ErrorInfo{}
		}

		if info.Message != "" {
			err.Message = info.Message
		}

		if info.Cause != "" {
			err.Cause = info.Cause
		}

		if info.Detail != nil {
			err.Detail = info.Detail
		}
		if info.Description != "" {
			err.Description = info.Description
		}
		if info.Solution != "" {
			err.Solution = info.Solution
		}

		return &err
	}
}

// buildCode根据传入的HTTP状态码，服务标识和错误标识，组合Error对象中的Code
//
// example: buildCode(PolicyManagement, http.StatusBadRequest, 1) // 400013001
func buildCode(serviceCode int, statusCode, errorCode int) int {
	return statusCode*1e6 + serviceCode*1e3 + errorCode
}

/*
NewErrorBuilder 是一个高阶函数，返回一个错误对象的生成器

serviceCode 为服务标识，需要先在Confluence《AS7服务定义》上进行登记，并在service.go中提交

示例：

// 创建一个策略服务的错误生成器
ErrorBuilder := NewErrorBuilder(PolicyManagement)

// 生成一个表示用户不存在的错误函数
ErrUserNotFound := ErrorBuilder.OfType(&ErrorType{
	StatusCode: http.StatusNotFound,
	ErrorCode:  1,
	ErrorInfo: ErrorInfo{
		Message: "用户不存在",
		Cause:   "数据库中未查找到对应用户id",
	},
})

// 返回具体的错误对象
ErrUserNotFound(&ErrorInfo{
	Detail: map[string]string{
		"userid": "151bcb65-48ce-4b62-973f-0bb6685f9cb8",
	},
}))
*/
func NewErrorBuilder(serviceCode int) TypedErrorBuilder {
	return TypedErrorBuilder{
		serviceCode: serviceCode,
		usedCode:    map[int]bool{},
	}
}
