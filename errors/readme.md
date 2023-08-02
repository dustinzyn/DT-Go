# 错误码使用说明

## 错误响应体数据结构
``` golang
type ErrorResp struct {
	code        int
	message     string
	cause       string
	detail      interface{}
	description string
	solution    string
}
```
* `code` 错误码 前三位等于状态码，中间三位为服务标识，后三位为错误标识。
* `message` 为旧版本错误信息字段，为兼容旧版本所以保留。
* `description` 为新版本错误信息字段，需要做国际化。
* `solution` 表示出现此错误时，需要的提示信息，需要做国际化。
* `cause` 表示产生错误的具体原因，不需要国际化，可以使用引用方的错误信息
* `detail` 为错误码的扩展信息，一般为可序列化的数据结构

开放的接口如下

``` golang
type APIError interface {
	// Marshal 序列化错误
	Marshal() []byte
	// Codes 已使用的错误码
	Codes() []int
	// Code 错误码 前三位等于状态码，中间三位为服务标识，后三位为错误标识
	Code() int
	// Error 格式化错误信息
	Error() string
	// StatusCode 状态码
	StatusCode() int
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
```
### 两个重点方法说明
1. 通过`Localization`方法将客户定义的国际化资源注册进来
``` golang
// i18nSet key:9位的错误码  value: I18n
func Localization(i18nSet map[int]I18n)

// I18n
// Description和Solution数据结构：key:语言 value:对应语言的国际化信息
// eg. {"zh-CN": "错误的请求。", "zh-TW": "錯誤的請求。", "en-US": "Invalid request."}
type I18n struct {
	Description map[string]string // 错误信息
	Solution    map[string]string // 操作提示
}
```
2. 通过`New`方法生成一个错误码响应体
```golang
func New(language string, code int, cause string, detail interface{}) *ErrorResp
```
* language: 国际化对应的语言，支持三种："zh-CN", "zh-TW", "en-US"，不传默认为系统语言
* code: 错误码
* cause：产生错误的具体原因
* detail: 错误的扩展信息

## 框架内置的错误码
框架里内置了一部分通用的错误码，可以直接使用
```golang
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
```

## 如何自定义一个带有国际化信息的错误码响应体

以定义一个400的错误码为例

1. 定义错误码
```golang
const BadRequestErr = 400000000
```
2. 定义国际化信息
```golang
errorI18n = map[int]I18n{
		BadRequestErr: {
			Description: map[string]string{
				"zh-CN": "错误的请求。",
				"zh-TW": "錯誤的請求。",
				"en-US": "Invalid request.",
			},
			Solution: map[string]string{
				"zh-CN": "请检查请求参数是否正确",
				"zh-TW": "請檢查請求參數是否正確。",
				"en-US": "Please check the request parameters to ensure they are correct.",
			},
		},
	}
```
3. 注册定义好的国际化资源
```golang
import "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"

errors.Localization(errorI18n)
```
4. 使用错误码
```golang
import "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive/errors"

// 生成错误码对象
err = errors.New("", BadRequestErr, "", nil)
// 使用错误码信息
code := err.Code()
description := err.Description()
solution := err.Solution()
cause := err.Cause()
detail := errr.Detail()
```