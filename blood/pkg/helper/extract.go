package helper

import (
	"fmt"
	"regexp"
	"strings"
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
	text = strings.TrimSpace(text)
	if text == "" {
		return "", "", fmt.Errorf("invalid function call format: empty text")
	}

	open := strings.Index(text, "(")
	if open <= 0 {
		return "", "", fmt.Errorf("invalid function call format: missing '(' . Got: %s", text)
	}

	functionName = strings.TrimSpace(text[:open])
	if functionName == "" || !regexp.MustCompile(`^\w+$`).MatchString(functionName) {
		return "", "", fmt.Errorf("invalid function call format: invalid function name. Got: %s", text)
	}

	rest := strings.TrimSpace(text[open+1:])
	if rest == "" {
		return functionName, "", nil
	}

	if strings.HasSuffix(text, ")") {
		rest = strings.TrimSuffix(rest, ")")
	}
	return functionName, rest, nil
}

func StripFrontMatter(content string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "---") {
		return content
	}

	// 查找第二个 "---"
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return content
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx != -1 && endIdx < len(lines)-1 {
		return strings.TrimSpace(strings.Join(lines[endIdx+1:], "\n"))
	}

	return content
}
