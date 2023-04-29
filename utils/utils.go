// 通用工具函数
package utils

import (
	"math/rand"
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
