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
				ZHCN: "请求错误",
				ZHTW: "請求錯誤",
				ENUS: "Bad request",
			},
			Solution: map[string]string{
				ZHCN: "请稍后重试或联系管理员。",
				ZHTW: "請稍後重試或聯絡管理員。",
				ENUS: "You can try again later or contact Admin.",
			},
		},
		UnauthorizedErr: {
			Description: map[string]string{
				ZHCN: "授权错误",
				ZHTW: "授權錯誤",
				ENUS: "Unauthorized",
			},
			Solution: map[string]string{
				ZHCN: "请检查您的用户名或密码后重新登录。",
				ZHTW: "請檢查您的使用者名稱或密碼后重新登入。",
				ENUS: "Please check your username or password and try again.",
			},
		},
		ForbiddenErr: {
			Description: map[string]string{
				ZHCN: "禁止访问",
				ZHTW: "禁止存取",
				ENUS: "Forbidden",
			},
			Solution: map[string]string{
				ZHCN: "您没有访问此页面的权限。",
				ZHTW: "您沒有存取此頁面的權限。",
				ENUS: "Not allowed to access this page.",
			},
		},
		ResourceNotFoundErr: {
			Description: map[string]string{
				ZHCN: "请求的资源不存在。",
				ZHTW: "請求的資源不存在。",
				ENUS: "Requested resource does not exist..",
			},
			Solution: map[string]string{
				ZHCN: "请稍后重试或联系管理员。",
				ZHTW: "請稍後重試或聯絡管理員。",
				ENUS: "You can try again later or contact Admin.",
			},
		},
		MethodNotAllowedErr: {
			Description: map[string]string{
				ZHCN: "请求方法错误",
				ZHTW: "請求方法錯誤",
				ENUS: "Method error",
			},
			Solution: map[string]string{
				ZHCN: "请稍后重试。如果长时间仍无反应，您可以联系管理员。",
				ZHTW: "請稍後重試。如果長時間仍無反應，您可以聯絡管理員。",
				ENUS: "You can try again later or contact Admin.",
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
				ZHCN: "当前服务器请求冲突",
				ZHTW: "當前伺服器請求衝突",
				ENUS: "Conflict detected in the server",
			},
			Solution: map[string]string{
				ZHCN: "您可以稍后重试。如果长时间仍无反应，您可以联系管理员。",
				ZHTW: "您可以稍後重試。如果長時間仍無反應，您可以聯絡管理員。",
				ENUS: "You can try again later or contact Admin.",
			},
		},
		InternalErr: {
			Description: map[string]string{
				ZHCN: "当前服务器存在内部错误或错误配置",
				ZHTW: "當前伺服器存在內部錯誤或錯誤設定",
				ENUS: "An internal error or wrong configuration has been detected",
			},
			Solution: map[string]string{
				ZHCN: "您可以稍后重试或联系管理员。",
				ZHTW: "您可以稍後重試或聯絡管理員。",
				ENUS: "You can try again later or contact Admin.",
			},
		},
	}
)
