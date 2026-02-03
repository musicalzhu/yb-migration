package report

import (
	"fmt"
	"path/filepath"
	"strings"
)

// validateOutputPath 验证输出路径的安全性
func validateOutputPath(path string) error {
	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查路径遍历攻击
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("路径包含上级目录访问")
	}

	// 检查绝对路径（如果需要限制）
	if filepath.IsAbs(cleanPath) {
		// 可以根据需要决定是否允许绝对路径
		// 这里我们允许，但可以添加额外的验证
		_ = cleanPath // 避免空块警告
	}

	// 检查危险字符（排除 Windows 驱动器路径中的冒号）
	dangerousChars := []string{"<", ">", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		if strings.Contains(cleanPath, char) {
			return fmt.Errorf("路径包含危险字符: %s", char)
		}
	}

	// 特殊检查：允许 Windows 驱动器路径中的冒号
	// 例如：C:\path\to\file.md 是合法的
	if strings.Contains(cleanPath, ":") {
		// 检查是否是合法的 Windows 路径格式
		// 冒号后面应该是反斜杠或正斜杠
		colonIndex := strings.Index(cleanPath, ":")
		if colonIndex > 0 && colonIndex < len(cleanPath)-1 {
			nextChar := cleanPath[colonIndex+1]
			if nextChar != '\\' && nextChar != '/' {
				return fmt.Errorf("路径包含危险字符: %s", ":")
			}
		}
	}

	return nil
}
