// 通用工具函数
package utils

import (
	"math/rand"
	"os"
	"time"
)

const (
	letters string = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~-_."
)

func RandString(length int) string {
	// 生成[6, 100]以内的随机数
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
