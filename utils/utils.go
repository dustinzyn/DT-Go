// 通用工具函数
package utils

import (
	"database/sql"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	dt "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/DT-Go"
)

const (
	letters string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~-_."
)

// RandString 生成指定范围内的随机数
func RandString(length int) string {
	letter := []rune(letters)
	b := make([]rune, length)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letter[r.Intn(len(letter))]
	}
	return string(b)

}

// 判断a、b两个数组是否有交集
func HasIntersection(a, b []string) bool {
	m := make(map[string]bool)
	for _, x := range a {
		m[x] = true
	}
	for _, x := range b {
		if m[x] {
			return true
		}
	}
	return false
}

// GetEnv 封装os.Getenv(),可以指定默认值
func GetEnv(key, defaultV string) string {
	v := os.Getenv(key)
	if v == "" {
		v = defaultV
	}
	return v
}

// StrToInt 将字符串转换为int
func StrToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// StrToUint64 将字符串转为Uint64
func StrToUint64(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}

// IntToStr int转string
func IntToStr(i int) string {
	return strconv.Itoa(i)
}

// Uint64ToStr uint64转string
func Uint64ToStr(i uint64) string {
	return strconv.FormatUint(i, 10)
}

// GetDefaultLanguage 获取系统默认语言
func GetDefaultLanguage() string {
	lang := GetEnv("LANGUAGE", "zh_CN")
	// 系统里定义的语言格式和http语言格式有差异
	lang = strings.ReplaceAll(lang, "_", "-")
	return lang
}

// ParseXLanguage 解析 http headers x-Language
func ParseXLanguage(xLanguage string, acceptLangs ...string) (language string) {
	// eg. zh-CH, fr;q=0.9, en-US;q=0.8, de;q=0.7, *;q=0.5
	xLanguage = strings.ReplaceAll(xLanguage, " ", "")
	// 支持的语言集
	acceptLangMap := make(map[string]string)
	if len(acceptLangs) != 0 {
		for _, lang := range acceptLangs {
			acceptLangMap[lang] = "1"
		}
	} else {
		acceptLangMap = map[string]string{
			"zh-CN": "1",
			"zh-TW": "1",
			"en-US": "1",
		}
	}

	defer func() {
		if r := recover(); r != nil {
			dt.Logger().Errorf("parse x-language: %v, error: %v", xLanguage, r)
			language = GetDefaultLanguage()
			return
		}
		if language == "" {
			language = GetDefaultLanguage()
			return
		}
	}()

	langWeights := strings.Split(xLanguage, ",")
	langWeightMap := make(map[string]string)
	for _, langw := range langWeights {
		langWeight := strings.Split(langw, ";")
		switch len(langWeight) {
		case 1:
			// [zh-CH]
			langWeightMap[langWeight[0]] = "1"
		case 2:
			// [fr,q=0.9]
			weight := strings.Split(langWeight[1], "=")[1]
			langWeightMap[langWeight[0]] = weight
		default:
		}
	}
	acceptWeight := ""
	acceptAll := false
	for lang, weight := range langWeightMap {
		// 命中已支持的语言集，并且权重最高的
		if _, ok := acceptLangMap[lang]; ok {
			if weight > acceptWeight {
				acceptWeight = weight
				language = lang
			}
		}
		if lang == "*" {
			acceptAll = true
		}
	}
	// 未命中但存在通配符* 采用默认语言
	if language == "" && acceptAll {
		language = GetDefaultLanguage()
	}
	// TODO 都未命中考虑降级，语言标签去掉区域后再匹配

	return
}

// CloseRows closes the Rows, preventing further enumeration. If Next is called
// and returns false and there are no further result sets,
// the Rows are closed automatically and it will suffice to check the
// result of Err. Close is idempotent and does not affect the result of Err.
func CloseRows(rows *sql.Rows) {
	if rows != nil {
		if rowsErr := rows.Err(); rowsErr != nil {
			dt.Logger().Error(rowsErr)
		}

		if closeErr := rows.Close(); closeErr != nil {
			dt.Logger().Error(closeErr)
		}
	}
}

// NowTimestamp 获取当前时间戳 毫秒级13位
func NowTimestamp() int64 {
	now := time.Now()
	return now.UnixNano() / 1000
}
