package errors

func init() {
	Localization(errorI18n)
}

var (
	// ZHCN 中文简体
	ZHCN = SupportLanguages[0]
	// ZHTW 中文繁体
	ZHTW = SupportLanguages[1]
	// ENUS 英文美国
	ENUS = SupportLanguages[2]

	errorI18n = map[int]I18n{
		BadRequestErr: {
			Description: map[string]string{
				ZHCN: "错误的请求。",
				ZHTW: "錯誤的請求。",
				ENUS: "Invalid request.",
			},
			Solution: map[string]string{
				ZHCN: "请检查请求参数是否正确",
				ZHTW: "請檢查請求參數是否正確。",
				ENUS: "Please check the request parameters to ensure they are correct.",
			},
		},
		InternalErr: {
			Description: map[string]string{
				ZHCN: "服务器内部错误。",
				ZHTW: "服務器內部錯誤。",
				ENUS: "Server internal error.",
			},
			Solution: map[string]string{
				ZHCN: "请联系管理员或刷新页面。",
				ZHTW: "請聯繫管理員或刷新頁面。",
				ENUS: "Please contact the administrator or refresh the page.",
			},
		},
		UnauthorizedErr: {
			Description: map[string]string{
				ZHCN: "请求未经授权。",
				ZHTW: "請求未經授權。",
				ENUS: "Request unauthorized.",
			},
			Solution: map[string]string{
				ZHCN: "请刷新页面更新token或重新登录。",
				ZHTW: "請刷新頁面更新token或重新登錄。",
				ENUS: "Please refresh the page to update the token or log in again.",
			},
		},
		ResourceNotFoundErr: {
			Description: map[string]string{
				ZHCN: "请求的资源不存在。",
				ZHTW: "請求的資源不存在。",
				ENUS: "Requested resource does not exist..",
			},
			Solution: map[string]string{
				ZHCN: "请检查请求的地址是否正确或联系管理员。",
				ZHTW: "檢查請求的地址是否正確或聯繫管理員。",
				ENUS: "Please check the requested address is correct or contact the administrator.",
			},
		},
		ForbiddenErr: {
			Description: map[string]string{
				ZHCN: "请求被拒绝。",
				ZHTW: "請求被拒絕。",
				ENUS: "Request rejected.",
			},
			Solution: map[string]string{
				ZHCN: "请联系管理员或重新登录。",
				ZHTW: "請聯繫管理員或重新登錄。",
				ENUS: "Please contact the administrator or log in again.",
			},
		},
		ConflictErr: {
			Description: map[string]string{
				ZHCN: "资源冲突。",
				ZHTW: "資源衝突。",
				ENUS: "Resource confict.",
			},
			Solution: map[string]string{
				ZHCN: "请联系管理员或刷新页面。",
				ZHTW: "請聯繫管理員或刷新頁面。",
				ENUS: "Please contact the administrator or refresh the page.",
			},
		},
		TooManyRequestsErr: {
			Description: map[string]string{
				ZHCN: "请求过于频繁。",
				ZHTW: "請求過於頻繁。",
				ENUS: "Too many requests.",
			},
			Solution: map[string]string{
				ZHCN: "请刷新页面后再重试。",
				ZHTW: "請刷新頁面後再重試。",
				ENUS: "Please refresh the page and try again.",
			},
		},
	}
)
