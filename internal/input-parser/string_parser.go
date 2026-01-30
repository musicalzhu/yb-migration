package inputparser

import (
	"fmt"
)

// StringParser 处理字符串输入
type StringParser struct{}

// NewStringParser 创建并返回一个新的字符串解析器
func NewStringParser() *StringParser {
	return &StringParser{}
}

// Parse 解析字符串中的SQL语句
// 参数 content 是包含SQL语句的字符串
// 返回值: 输入字符串和可能的错误
func (p *StringParser) Parse(content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("输入内容不能为空")
	}

	return content, nil
}
