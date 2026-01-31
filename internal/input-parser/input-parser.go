package inputparser

import (
	"io"
	"strings"
)

// InputParser 定义输入解析器接口
type InputParser interface {
	// Parse 从输入源解析内容并返回SQL语句字符串
	// 参数 path 是文件路径或SQL字符串
	Parse(path string) (string, error)
}

// InputType 表示输入源类型
type InputType string

const (
	// WindowsMaxPath 是 Windows 系统下的最大文件路径长度
	WindowsMaxPath = 260

	// InputTypeSQLFile 表示以文件路径形式的 SQL 文件输入类型。
	InputTypeSQLFile    InputType = "sqlFile"    // SQL文件
	// InputTypeGeneralLog 表示 MySQL general log 文件输入类型（例如 slow log / general log）。
	InputTypeGeneralLog InputType = "generalLog" // MySQL general log文件
	// InputTypeString 表示直接传入的 SQL 字符串输入类型。
	InputTypeString     InputType = "string"     // 字符串
	// InputTypeStream 表示基于 io.Reader 的流式输入类型。
	InputTypeStream     InputType = "stream"     // 流输入
)

// NewParser 根据输入类型创建相应的解析器
func NewParser(inputType InputType) InputParser {
	switch inputType {
	case InputTypeSQLFile:
		return NewSQLFileParser()
	case InputTypeGeneralLog:
		return NewGeneralLogFileParser()
	case InputTypeString:
		return NewStringParser()
	default:
		return nil
	}
}

// isPotentialFilePath 检查字符串是否可能是文件路径
func isPotentialFilePath(s string) bool {
	if len(s) >= WindowsMaxPath {
		return false
	}

	// 定义支持的文件扩展名
	supportedExts := []string{".sql", ".log"}
	lowerStr := strings.ToLower(s)

	for _, ext := range supportedExts {
		if strings.HasSuffix(lowerStr, ext) {
			return true
		}
	}
	return false
}

// DetectInputType 检测输入源类型
// 返回输入源类型，如果无法确定则返回空字符串
func DetectInputType(source any) InputType {
	switch v := source.(type) {
	case string:
		if isPotentialFilePath(v) {
			// 根据文件扩展名返回具体类型
			lowerStr := strings.ToLower(v)
			if strings.HasSuffix(lowerStr, ".sql") {
				return InputTypeSQLFile
			} else if strings.HasSuffix(lowerStr, ".log") {
				return InputTypeGeneralLog
			}
		}
		return InputTypeString
	case io.Reader:
		return InputTypeStream
	}
	return ""
}
