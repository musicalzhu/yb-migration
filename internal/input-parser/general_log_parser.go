// Package inputparser 提供对不同输入类型（如 SQL 文件、MySQL general log 等）的解析器实现。
// 各解析器负责将原始输入转换为可供分析的 SQL 文本。
package inputparser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GeneralLogFileParser 专门用于解析MySQL general log文件
// 支持从MySQL general log中提取SQL查询语句
// 注意：此解析器仅处理标准格式的MySQL general log
type GeneralLogFileParser struct {
	// 匹配MySQL general log中的行
	// 格式示例: 2023-12-23T08:00:01.234567Z     1 Query     SELECT * FROM users
	// 组1: 完整匹配
	// 组2: 时间戳
	// 组3: 线程ID
	// 组4: 命令类型(Query, Connect等)
	// 组5: SQL语句
	logLinePattern *regexp.Regexp
	// 存储非标准格式的日志行
	nonStandardLines []string
}

// NewGeneralLogFileParser 创建并初始化一个新的MySQL general log解析器
func NewGeneralLogFileParser() *GeneralLogFileParser {
	pattern := `^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+(\d+)\s+(\w+)\s+(.*)$`

	return &GeneralLogFileParser{
		logLinePattern:   regexp.MustCompile(pattern),
		nonStandardLines: []string{},
	}
}

// GetNonStandardLines 获取所有非标准格式的日志行
func (p *GeneralLogFileParser) GetNonStandardLines() []string {
	return p.nonStandardLines
}

// Parse 解析MySQL general log文件
// 参数 path 是日志文件的路径
// 返回值: 提取的SQL语句字符串和可能的错误
// 注意：不支持目录，请使用 analyzer 的 AnalyzeInput 方法处理目录
func (p *GeneralLogFileParser) Parse(path string) (string, error) {
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
		return "", fmt.Errorf("不支持目录，请使用 analyzer 的 AnalyzeInput 方法")
	}

	// 检查文件扩展名
	if !isLogFile(path) {
		return "", fmt.Errorf("不支持的文件类型: %s，日志文件通常为 .log 扩展名", filepath.Ext(path))
	}

	// 读取文件内容
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开文件 %s 失败: %w", path, err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭文件 %s 失败: %v\n", path, err)
		}
	}()

	return p.parseGeneralLog(file)
}

// parseGeneralLog 从io.Reader中解析general log内容
func (p *GeneralLogFileParser) parseGeneralLog(reader io.Reader) (string, error) {
	var sqlContent strings.Builder
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		sql, err := p.parseLogLine(line)
		if err != nil {
			return "", err
		}
		if sql != "" {
			sqlContent.WriteString(sql)
			// 添加分号和换行符，确保多条SQL语句正确分隔
			if !strings.HasSuffix(sql, ";") {
				sqlContent.WriteString(";")
			}
			sqlContent.WriteString("\n")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取日志内容时出错: %w", err)
	}

	return sqlContent.String(), nil
}

// parseLogLine 解析单行日志
// 返回:
//   - 解析出的SQL语句
//   - 错误信息
//     注意：非标准格式的日志行会被记录到 nonStandardLines 中
func (p *GeneralLogFileParser) parseLogLine(line string) (string, error) {
	// 跳过空行
	line = strings.TrimSpace(line)
	if line == "" {
		return "", nil
	}

	// 匹配日志行格式
	matches := p.logLinePattern.FindStringSubmatch(line)
	if len(matches) < 5 {
		// 记录非标准格式的日志行
		p.nonStandardLines = append(p.nonStandardLines, line)
		return "", nil
	}

	// 获取命令类型
	commandType := matches[3]

	// 只处理Query类型的日志
	if commandType != "Query" {
		// 记录非Query类型的日志行
		p.nonStandardLines = append(p.nonStandardLines, fmt.Sprintf("[Non-Query] %s", line))
		return "", nil
	}

	// 提取并清理SQL语句
	sql := strings.TrimSpace(matches[4])
	if sql == "" || isIgnoredSQL(sql) {
		return "", nil
	}

	return sql, nil
}

// isLogFile 检查文件是否为日志文件
// 参数:
//   - path: 文件路径
//
// 返回值:
//   - bool: true表示是日志文件，false表示不是
//
// 说明:
//
//	通过文件扩展名判断，仅支持 .log 文件
func isLogFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".log"
}

// isIgnoredSQL 检查是否是需要忽略的SQL语句
// 参数:
//   - sql: SQL语句字符串
//
// 返回值:
//   - bool: true表示需要忽略，false表示需要处理
//
// 实现细节:
//  1. 转换为大写进行不区分大小写比较
//  2. 检查SQL前缀是否匹配忽略列表
//
// 忽略的SQL类型:
//   - SET语句: 设置变量
//   - SHOW语句: 显示信息
//   - USE语句: 切换数据库
//   - 事务控制: BEGIN, COMMIT, ROLLBACK
//   - 系统查询: SELECT DATABASE(), SELECT USER(), SELECT @@
func isIgnoredSQL(sql string) bool {
	// 转换为大写以进行不区分大小写的比较
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))

	// 需要忽略的SQL类型
	ignoredPrefixes := []string{
		"SET ", "SHOW ", "USE ", "BEGIN", "COMMIT",
		"ROLLBACK", "START TRANSACTION", "SET NAMES",
		"SELECT DATABASE()", "SELECT USER()", "SELECT @@",
	}

	for _, prefix := range ignoredPrefixes {
		if strings.HasPrefix(upperSQL, prefix) {
			return true
		}
	}

	return false
}
