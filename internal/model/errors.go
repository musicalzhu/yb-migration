// Package model 定义了分析过程中使用的错误类型和常量
package model

import (
	"fmt"
)

// ErrorType 定义了错误的类型
type ErrorType string

const (
	// ErrorTypeParse SQL 解析错误
	ErrorTypeParse ErrorType = "parse"
	// ErrorTypeNoSQL 未找到有效 SQL 错误
	ErrorTypeNoSQL ErrorType = "no_sql"
	// ErrorTypeTransform SQL 转换错误
	ErrorTypeTransform ErrorType = "transform"
	// ErrorTypeConfig 配置错误
	ErrorTypeConfig ErrorType = "config"
	// ErrorTypeFile 文件操作错误
	ErrorTypeFile ErrorType = "file"
)

// AnalysisError 表示分析过程中发生的错误
type AnalysisError struct {
	Type    ErrorType `json:"type"`
	Message string    `json:"message"`
	Source  string    `json:"source"`
	SQL     string    `json:"sql,omitempty"`
	Line    int       `json:"line,omitempty"`
	Column  int       `json:"column,omitempty"`
	Cause   error     `json:"-"` // 原始错误，不序列化
}

// Error 实现 error 接口
func (e *AnalysisError) Error() string {
	if e.Source != "" {
		return fmt.Sprintf("[%s] %s (source: %s)", e.Type, e.Message, e.Source)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap 支持错误链
func (e *AnalysisError) Unwrap() error {
	return e.Cause
}

// Is 支持错误类型比较
func (e *AnalysisError) Is(target error) bool {
	if t, ok := target.(*AnalysisError); ok {
		return e.Type == t.Type
	}
	return false
}

// 预定义的错误变量，用于 errors.Is 检查
var (
	// ErrParse 解析错误
	ErrParse = &AnalysisError{Type: ErrorTypeParse, Message: "SQL解析失败"}
	// ErrNoSQL 无SQL错误
	ErrNoSQL = &AnalysisError{Type: ErrorTypeNoSQL, Message: "未找到有效SQL"}
	// ErrTransform 转换错误
	ErrTransform = &AnalysisError{Type: ErrorTypeTransform, Message: "SQL转换失败"}
	// ErrConfig 配置错误
	ErrConfig = &AnalysisError{Type: ErrorTypeConfig, Message: "配置错误"}
	// ErrFile 文件错误
	ErrFile = &AnalysisError{Type: ErrorTypeFile, Message: "文件操作错误"}
)
