package helper

import (
	"fmt"
	"regexp"
)

func ExtractContentByTag(text, tagName string) string {
	// 构建正则表达式，(?s) 标志让 . 可以匹配换行符
	// (.*?) 是一个非贪婪的捕获组，用于匹配标签之间的任何内容
	pattern := fmt.Sprintf(`(?s)<%s>(.*?)</%s>`, regexp.QuoteMeta(tagName), regexp.QuoteMeta(tagName))

	// 编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		// 通常，动态构建的正则表达式不会出错，但最好还是处理一下
		fmt.Println("Error compiling regex:", err)
		return ""
	}

	// 查找子匹配项
	matches := re.FindStringSubmatch(text)

	// FindStringSubmatch 返回一个切片，其中：
	// - 第0个元素是整个匹配的字符串（例如 "<tag>content</tag>"）
	// - 第1个元素是第一个捕获组的内容（即我们想要的内容）
	if len(matches) > 1 {
		return matches[1]
	}

	// 如果没有找到匹配项
	return ""
}

func ParseFunctionCall(text string) (functionName string, args string, err error) {
	// 正则表达式解析：
	// ^(\w+)       - 匹配并捕获开头的函数名（字母、数字、下划线）。
	// \s*          - 匹配函数名和括号之间可能存在的任意空格。
	// \(           - 匹配一个字面的左括号。
	// (.*)         - 捕获括号内的所有内容。
	// \)           - 匹配一个字面的右括号。
	// $            - 确保匹配到字符串末尾。
	re := regexp.MustCompile(`^(\w+)\s*\((.*)\)$`)

	// 查找匹配项和子匹配（捕获组）
	matches := re.FindStringSubmatch(text)

	// FindStringSubmatch 应该返回3个元素：
	// 1. 整个匹配的字符串
	// 2. 第一个捕获组（函数名）
	// 3. 第二个捕获组（参数）
	if len(matches) != 3 {
		return "", "", fmt.Errorf("invalid function call format: text does not match 'function(...)'. Got: %s", text)
	}

	functionName = matches[1]
	args = matches[2]
	err = nil

	return
}
