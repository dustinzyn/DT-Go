package errors

import (
	"encoding/json"
	"fmt"
	"strings"

	"DT-Go/utils"
)

var (
	// i18ns 已注册的错误码国际化信息
	i18ns = make(map[int]I18n)
	// SupportLanguages 支持的语言
	SupportLanguages = [3]string{"zh-CN", "zh-TW", "en-US"}
)

// I18n 国际化资源
type I18n struct {
	Description map[string]string // 错误信息
	Solution    map[string]string // 操作提示
}

// Localization 提供所有语言的翻译和本地化
func Localization(i18nSet map[int]I18n) {
	for code := range i18nSet {
		if _, used := i18ns[code]; used {
			panic(fmt.Errorf("duplicate code: %v", code))
		}
		for _, lang := range SupportLanguages {
			if _, ok := i18nSet[code].Description[lang]; !ok {
				panic(fmt.Errorf("unsupport language in description."))
			}
			if _, ok := i18nSet[code].Solution[lang]; !ok {
				panic(fmt.Errorf("unsupport language in solution."))
			}
		}
		i18ns[code] = i18nSet[code]
	}
}

// APIError .
type APIError interface {
	// Marshal 序列化错误
	Marshal() []byte
	// Codes 已使用的错误码
	Codes() []int
	// Code 错误码 前三位等于状态码，中间三位为服务标识，后三位为错误标识
	Code() int
	// StatusCode 状态码
	StatusCode() int
	// Error 格式化错误信息
	Error() string
	// Message 错误信息
	Message() string
	// Cause 产生错误的具体原因
	Cause() string
	// Detail 补充说明错误信息
	Detail() interface{}
	// Description 错误描述 不设置默认使用message的值
	Description() string
	// Solution 操作提示
	Solution() string
}

// New 新建错误码返回体
func New(language string, code int, cause string, detail interface{}) *ErrorResp {
	// 未设置语言，使用系统默认语言
	if language == "" {
		language = utils.GetDefaultLanguage()
	}
	return &ErrorResp{
		code:        code,
		cause:       cause,
		detail:      detail,
		description: i18ns[code].Description[language],
		solution:    i18ns[code].Solution[language],
	}
}

// ErrorResp .
type ErrorResp struct {
	code        int         // 错误码 前三位等于状态码，中间三位为服务标识，后三位为错误标识
	message     string      // 错误信息，旧版的字段，可以不填，会沿用description的值
	cause       string      // 错误原因，产生错误的具体原因
	detail      interface{} // 错误码拓展信息，补充说明错误信息
	description string      // 错误描述，客户端采用此字段做错误提示（需要符合国际化要求）
	solution    string      // 操作提示，针对当前错误的操作提示（需要符合国际化要求）
}

// Marshal .
func (e *ErrorResp) Marshal() []byte {
	result := make(map[string]interface{})
	result["code"] = e.code
	result["cause"] = e.cause
	result["detail"] = e.detail
	result["description"] = e.description
	result["solution"] = e.solution
	resByte, _ := json.Marshal(&result)
	return resByte
}

// Codes 已注册的错误码
func (e *ErrorResp) Codes() []int {
	codes := make([]int, 0)
	for code := range i18ns {
		codes = append(codes, code)
	}
	return codes
}

// Code 错误码
func (e *ErrorResp) Code() int {
	return e.code
}

// StatusCode 状态码
func (e *ErrorResp) StatusCode() int {
	code := utils.IntToStr(e.Code())[:3]
	return utils.StrToInt(code)
}

// Message .
func (e *ErrorResp) Message() string {
	if e.message == "" {
		return e.description
	}
	return e.message
}

// Cause .
func (e *ErrorResp) Cause() string {
	return e.cause
}

// Detail .
func (e *ErrorResp) Detail() interface{} {
	return e.detail
}

// Description .
func (e *ErrorResp) Description() string {
	return e.description
}

// Solution .
func (e *ErrorResp) Solution() string {
	return e.solution
}

// Error .
func (e *ErrorResp) Error() string {
	errInfo := []string{}

	if e.Code() != 0 {
		errInfo = append(errInfo, fmt.Sprintf("Code: %d", e.Code()))
	}

	if e.description != "" {
		errInfo = append(errInfo, fmt.Sprintf("Description: %s", e.Description()))
	}

	if e.cause != "" {
		errInfo = append(errInfo, fmt.Sprintf("Cause: %s", e.Cause()))
	}

	if e.solution != "" {
		errInfo = append(errInfo, fmt.Sprintf("Solution: %s", e.Solution()))
	}

	return strings.Join(errInfo, ", ")
}
