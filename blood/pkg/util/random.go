package util

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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

func RandomIdWithPrefix(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, generateID())
}

func Float32SliceToString(s []float32) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range s {
		if i > 0 {
			sb.WriteString(",")
		}
		// 使用 32 位精度格式化，尽量保持紧凑
		sb.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
	}
	sb.WriteString("]")
	return sb.String()
}
