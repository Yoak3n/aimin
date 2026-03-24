package agent

import (
	"regexp"
	"strings"
)

func ParseToolCall(text string) (string, map[string]string, bool) {
	text = strings.TrimSpace(text)
	// 支持解析函数名和参数，忽略外层可能包裹的转义字符或多余空白
	// 使用 (?s) 使得 . 可以匹配换行符
	toolRe := regexp.MustCompile(`(?s)(?P<name>[a-zA-Z0-9_]+)\s*\((?P<args>.*?)\)`)

	toolMatch := toolRe.FindStringSubmatch(text)
	if len(toolMatch) < 3 {
		return "", nil, false
	}

	name := toolMatch[1]
	argsStr := toolMatch[2]

	// 解析参数: 支持 key="value", key='value' 和 key=value，使用 (?s) 使得点可以匹配换行符
	argRe := regexp.MustCompile(`(?s)(?P<key>[a-zA-Z0-9_]+)\s*=\s*(?:"(?P<v1>(?:\\"|[^"])*)"|'(?P<v2>(?:\\'|[^'])*)'|(?P<v3>[^,\)]*))`)
	argMatches := argRe.FindAllStringSubmatch(argsStr, -1)

	args := make(map[string]string)
	for _, match := range argMatches {
		if len(match) < 5 {
			continue
		}
		key := match[1]
		if match[2] != "" {
			v1 := match[2]
			// 还原 \n 为真实换行符
			v1 = strings.ReplaceAll(v1, `\n`, "\n")
			v1 = strings.ReplaceAll(v1, `\"`, `"`)
			v1 = strings.ReplaceAll(v1, `\\`, `\`)
			args[key] = v1
		} else if match[3] != "" {
			v2 := match[3]
			v2 = strings.ReplaceAll(v2, `\n`, "\n")
			v2 = strings.ReplaceAll(v2, `\'`, `'`)
			v2 = strings.ReplaceAll(v2, `\\`, `\`)
			args[key] = v2
		} else if match[4] != "" {
			args[key] = strings.TrimSpace(match[4])
		}
	}

	return name, args, true
}
