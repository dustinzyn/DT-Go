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
				ZHCN: "请求错误，请稍后重试。",
				ZHTW: "請求錯誤，請稍後重試。",
				ENUS: "Bad request. You can try again later.",
			},
			Solution: map[string]string{
				ZHCN: "",
				ZHTW: "",
				ENUS: "",
			},
		},
		UnauthorizedErr: {
			Description: map[string]string{
				ZHCN: "授权错误，请检查您的用户名或密码后重新登录。",
				ZHTW: "授權錯誤，請檢查您的使用者名稱或密碼后重新登入。",
				ENUS: "Unauthorized. Please check your username or password and try again.",
			},
			Solution: map[string]string{
				ZHCN: "",
				ZHTW: "",
				ENUS: "",
			},
		},
		ForbiddenErr: {
			Description: map[string]string{
				ZHCN: "您没有权限访问当前资源。",
				ZHTW: "您沒有權限存取當前資源。",
				ENUS: "You are not allowed to access this resource.",
			},
			Solution: map[string]string{
				ZHCN: "",
				ZHTW: "",
				ENUS: "",
			},
		},
		ResourceNotFoundErr: {
			Description: map[string]string{
				ZHCN: "当前资源已不存在，请稍后重试或联系管理员。",
				ZHTW: "當前資源已不存在，請稍後重試或聯絡管理員。",
				ENUS: "This resource doesn't exist. You can try again later or contact Admin.",
			},
			Solution: map[string]string{
				ZHCN: "",
				ZHTW: "",
				ENUS: "",
			},
		},
		MethodNotAllowedErr: {
			Description: map[string]string{
				ZHCN: "请求方法错误。请稍后重试。如果长时间仍无反应，您可以联系管理员。",
				ZHTW: "請求方法錯誤。請稍後重試。如果長時間仍無反應，您可以聯絡管理員。",
				ENUS: "Method error. You can try again later or contact Admin.",
			},
			Solution: map[string]string{
				ZHCN: "",
				ZHTW: "",
				ENUS: "",
			},
		},
		ConflictErr: {
			Description: map[string]string{
				ZHCN: "当前服务器请求冲突，您可以稍后重试。",
				ZHTW: "當前伺服器請求衝突，您可以稍後重試。",
				ENUS: "Conflict detected in the server. You can try again later.",
			},
			Solution: map[string]string{
				ZHCN: "",
				ZHTW: "",
				ENUS: "",
			},
		},
		TooManyRequestsErr: {
			Description: map[string]string{
				ZHCN: "请求过多，您可以稍后重试。",
				ZHTW: "請求過多，您可以稍後重試。",
				ENUS: "Too many requests. You can try again later.",
			},
			Solution: map[string]string{
				ZHCN: "",
				ZHTW: "",
				ENUS: "",
			},
		},
		InternalErr: {
			Description: map[string]string{
				ZHCN: "当前服务器存在内部错误或错误配置",
				ZHTW: "當前伺服器存在內部錯誤或錯誤設定",
				ENUS: "An internal error or wrong configuration has been detected",
			},
			Solution: map[string]string{
				ZHCN: "",
				ZHTW: "",
				ENUS: "",
			},
		},
	}
)
