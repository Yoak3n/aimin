package util

import (
	"fmt"
	"math/rand"
)

func generateID() string {
	// 实现ID生成逻辑
	str := randomString(10)
	return str
}

func randomString(length int) string {
	// 实现随机字符串生成逻辑
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func GenEventID(name string) string {
	return fmt.Sprintf("e-%s-%s", name, generateID())
}
