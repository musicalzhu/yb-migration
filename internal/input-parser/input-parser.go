// Package inputparser 提供输入解析器接口和实现，支持从不同来源（文件、字符串、流）解析 SQL 内容。
// 支持多种输入类型：
//   - SQL 文件：直接读取 .sql 文件内容
//   - 日志文件：从 MySQL general log 中提取 SQL 语句
//   - 字符串：直接传入 SQL 字符串
//   - 流输入：基于 io.Reader 的流式输入
package inputparser

// InputParser 定义输入解析器接口
type InputParser interface {
	// Parse 从输入源解析内容并返回SQL语句字符串
	// 参数 path 是文件路径或SQL字符串
	Parse(path string) (string, error)
}
