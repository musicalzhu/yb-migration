package inputparser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SQLFileParser 处理SQL文件输入
// 支持 .sql 文件
// 支持目录递归解析
// 注意：不实现文件分片处理，原因：
// 1. 现代服务器通常有足够的内存处理大文件
// 2. 分片处理可能破坏完整的SQL语句，导致解析错误
// 3. 保持代码简洁，将复杂的SQL解析逻辑交给专门的SQL解析器处理
type SQLFileParser struct {
	// 可以添加配置选项，如字符集、是否忽略错误等
}

// NewSQLFileParser 创建并返回一个新的SQL文件解析器
func NewSQLFileParser() *SQLFileParser {
	return &SQLFileParser{}
}

// Parse 解析SQL文件
// 参数 path 是SQL文件的路径
// 返回值: SQL内容字符串和可能的错误
// 注意：不支持目录，请使用 analyzer 的 AnalyzeDirectory 方法处理目录
func (p *SQLFileParser) Parse(path string) (string, error) {
	// 路径检查
	if path == "" {
		return "", fmt.Errorf("文件路径不能为空")
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("文件 %s 不存在: %w", path, err)
	}

	// 不支持目录
	if fileInfo.IsDir() {
		return "", fmt.Errorf("不支持目录，请使用 analyzer 的 AnalyzeDirectory 方法")
	}

	// 检查文件扩展名
	if ext := strings.ToLower(filepath.Ext(path)); ext != ".sql" {
		return "", fmt.Errorf("不支持的文件类型: %s，仅支持 .sql 文件", ext)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取文件 %s 失败: %w", path, err)
	}

	// 返回文件内容字符串
	return string(content), nil
}
