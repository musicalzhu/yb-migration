// Package analyzer 将输入解析为 SQL 文本并调用 SQL 解析器与检查器进行分析。
// 支持文件、目录和日志输入类型，并根据文件类型分派相应的解析器。
package analyzer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/ybMigration/internal/checker"
	"github.com/example/ybMigration/internal/config"
	inputparser "github.com/example/ybMigration/internal/input-parser"
	"github.com/example/ybMigration/internal/model"
	sqlparser "github.com/example/ybMigration/internal/sql-parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
)

// ============================================================================
// SQL 分析器
// ============================================================================

// SQLAnalyzer SQL 分析器
type SQLAnalyzer struct {
	inputParser inputparser.InputParser
	sqlParser   sqlparser.SQLParser
	checkers    []checker.Checker
}

// NewSQLAnalyzer 创建 SQL 分析器
func NewSQLAnalyzer(inputParser inputparser.InputParser, sqlParser sqlparser.SQLParser, checkers []checker.Checker) (*SQLAnalyzer, error) {
	// 支持空的 checkers 列表（表示无需执行任何检查，完全兼容）
	// 直接返回 analyzer，不将其视为错误
	return &SQLAnalyzer{
		inputParser: inputParser,
		sqlParser:   sqlParser,
		checkers:    checkers,
	}, nil
}

// AnalyzeSQL 分析 SQL 语句字符串（支持转换）。
// 参数:
//   - sql: 要分析的 SQL 语句字符串
//   - source: SQL 来源标识（文件路径、输入字符串等）
//
// 返回值:
//   - model.AnalysisResult: 分析结果，包含问题列表和转换后的 SQL
//   - error: 如果发生错误，返回具体的 AnalysisError
//
// 错误处理策略:
//   - SQL 解析失败：返回 ErrorTypeParse 类型的 AnalysisError
//   - 未找到有效 SQL：返回 ErrorTypeNoSQL 类型的 AnalysisError
//   - 生成转换 SQL 失败：返回 ErrorTypeTransform 类型的 AnalysisError
//   - 分析过程中的兼容性问题：记录在 result.Issues 中，不返回 error
//
// 使用示例：
//
//	result, err := analyzer.AnalyzeSQL("SELECT * FROM users", "test.sql")
//	if err != nil {
//		var analysisErr *AnalysisError
//		if errors.As(err, &analysisErr) {
//			switch analysisErr.Type {
//			case ErrorTypeParse:
//				fmt.Printf("SQL 解析失败: %v\n", analysisErr)
//			case ErrorTypeNoSQL:
//				fmt.Printf("未找到有效 SQL: %v\n", analysisErr)
//			case ErrorTypeTransform:
//				fmt.Printf("SQL 转换失败: %v\n", analysisErr)
//			}
//		}
//		return
//	}
func (a *SQLAnalyzer) AnalyzeSQL(sql string, source string) (model.AnalysisResult, error) {
	// 解析 SQL 语句
	stmts, err := a.sqlParser.ParseSQL(sql)
	if err != nil {
		return model.AnalysisResult{
				SQL:    sql,
				Source: source,
			}, &model.AnalysisError{
				Type:    model.ErrorTypeParse,
				Message: "SQL 解析失败",
				Source:  source,
				SQL:     sql,
				Cause:   err,
			}
	}

	if len(stmts) == 0 {
		return model.AnalysisResult{
				SQL:    sql,
				Source: source,
			}, &model.AnalysisError{
				Type:    model.ErrorTypeNoSQL,
				Message: "未找到有效的 SQL 语句",
				Source:  source,
				SQL:     sql,
			}
	}

	// 使用 checker.Check 进行一次遍历完成分析和转换
	checkResult := checker.Check(stmts, a.checkers...)

	// 生成转换后的SQL
	transformedSQL, err := a.generateSQL(checkResult.TransformedStmts)
	if err != nil {
		return model.AnalysisResult{
				SQL:    sql,
				Source: source,
				Issues: checkResult.Issues,
			}, &model.AnalysisError{
				Type:    model.ErrorTypeTransform,
				Message: "生成转换SQL失败",
				Source:  source,
				SQL:     sql,
				Cause:   err,
			}
	}

	return model.AnalysisResult{
		SQL:            sql,
		Source:         source,
		Issues:         checkResult.Issues,
		TransformedSQL: transformedSQL,
	}, nil
}

// generateSQL 从AST节点生成SQL字符串（使用TiDB的Restore功能）
// 参数:
//   - stmts: AST语句节点列表
//
// 返回值:
//   - string: 生成的SQL字符串
//   - error: 生成过程中的错误
//
// 实现细节:
//  1. 使用TiDB的Restore API将AST转换回SQL
//  2. 使用strings.Builder提高字符串拼接性能
//  3. 添加RestoreStringWithoutCharset标志去除字符集前缀
//  4. 预分配slice容量避免多次扩容
//
// 注意事项:
//   - 空语句列表返回空字符串
//   - nil语句会被跳过
//   - 语句间用分号和换行符分隔
func (a *SQLAnalyzer) generateSQL(stmts []ast.StmtNode) (string, error) {
	if len(stmts) == 0 {
		return "", nil
	}

	// 使用TiDB的Restore API生成SQL
	// 预分配容量，提高性能
	sqlParts := make([]string, 0, len(stmts))
	for _, stmt := range stmts {
		if stmt != nil {
			// 创建strings.Builder作为RestoreWriter，避免字符串拼接性能问题
			var builder strings.Builder

			// 创建RestoreCtx，配置恢复标志
			// DefaultRestoreFlags = RestoreStringSingleQuotes | RestoreKeyWordUppercase | RestoreNameBackQuotes
			// 添加 RestoreStringWithoutCharset 标志来去除 _UTF8MB4 等字符集前缀
			// 移除 RestoreNameBackQuotes 标志，避免自动添加反引号
			flags := format.RestoreStringSingleQuotes | format.RestoreKeyWordUppercase | format.RestoreStringWithoutCharset
			ctx := format.NewRestoreCtx(flags, &builder)

			// 调用TiDB的Restore方法生成SQL
			// 这里可能会因为AST节点损坏而失败，需要错误处理
			err := stmt.Restore(ctx)
			if err != nil {
				// Restore失败时返回错误，不使用panic
				return "", fmt.Errorf("TiDB Restore失败: %w", err)
			}

			sqlStr := builder.String()
			if sqlStr != "" {
				sqlParts = append(sqlParts, sqlStr)
			}
		}
	}

	// 用分号和换行符连接多个SQL语句
	return strings.Join(sqlParts, ";\n"), nil
}

// isSupportedFileExt 检查文件扩展名是否受支持
// 参数:
//   - ext: 文件扩展名（包含点号，如 ".sql"）
//
// 返回值:
//   - bool: true表示支持，false表示不支持
//
// 说明:
//
//	支持的文件类型: .sql, .log
func isSupportedFileExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".sql", ".log":
		return true
	default:
		return false
	}
}

// newFileParserForExt 根据文件扩展名创建对应的解析器
// 参数:
//   - ext: 文件扩展名（包含点号，如 ".sql"）
//
// 返回值:
//   - inputparser.InputParser: 对应的输入解析器实例
//   - error: 不支持的文件类型时返回错误
//
// 说明:
//
//	.sql 文件 -> SQLFileParser
//	.log 文件 -> GeneralLogFileParser
func newFileParserForExt(ext string) (inputparser.InputParser, error) {
	switch strings.ToLower(ext) {
	case ".sql":
		return inputparser.NewSQLFileParser(), nil
	case ".log":
		return inputparser.NewGeneralLogFileParser(), nil
	default:
		return nil, fmt.Errorf("不支持的文件类型: %s，仅支持 .sql 和 .log 文件", ext)
	}
}

// ============================================================================
// 便捷函数
// ============================================================================

// ============================================================================
// 分析器工厂
// ============================================================================

// Factory 分析器工厂。保留原始名称 AnalyzerFactory 可在历史记录中参考。
type Factory struct {
	config *config.Config
}

// NewAnalyzerFactory 创建分析器工厂
func NewAnalyzerFactory(configPath string) (*Factory, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	return &Factory{
		config: cfg,
	}, nil
}

// CreateCheckers 创建检查器列表
func (f *Factory) CreateCheckers(categories ...string) ([]checker.Checker, error) {
	var checkers []checker.Checker

	// 如果未指定类别，返回空切片（表示不启用任何检查器）
	if len(categories) == 0 {
		return []checker.Checker{}, nil
	}

	for _, category := range categories {
		switch strings.ToLower(category) {
		case "datatype":
			dataTypeChecker, err := checker.NewDataTypeChecker(f.config)
			if err != nil {
				return nil, fmt.Errorf("创建数据类型检查器失败: %w", err)
			}
			checkers = append(checkers, dataTypeChecker)
		case "function":
			functionChecker, err := checker.NewFunctionChecker(f.config)
			if err != nil {
				return nil, fmt.Errorf("创建函数检查器失败: %w", err)
			}
			checkers = append(checkers, functionChecker)
		case "syntax":
			syntaxChecker, err := checker.NewSyntaxChecker(f.config)
			if err != nil {
				return nil, fmt.Errorf("创建语法检查器失败: %w", err)
			}
			checkers = append(checkers, syntaxChecker)
		case "charset":
			charsetChecker, err := checker.NewCharsetChecker(f.config)
			if err != nil {
				return nil, fmt.Errorf("创建字符集检查器失败: %w", err)
			}
			checkers = append(checkers, charsetChecker)
		default:
			return nil, fmt.Errorf("不支持的检查器类别: %s", category)
		}
	}

	return checkers, nil
}

// CreateCheckersFromConfig 根据已加载的配置自动创建对应的检查器
// 返回:
//   - []checker.Checker: 根据配置中 rules 的 category 创建的检查器列表
//   - error: 错误信息
func (f *Factory) CreateCheckersFromConfig() ([]checker.Checker, error) {
	categories := f.extractCategoriesFromConfig()
	return f.CreateCheckers(categories...)
}

// extractCategoriesFromConfig 从配置中提取所有规则类别
// 返回:
//   - []string: 去重后的类别列表
func (f *Factory) extractCategoriesFromConfig() []string {
	categorySet := make(map[string]struct{})
	for _, rule := range f.config.Rules {
		categorySet[rule.Category] = struct{}{}
	}

	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}
	return categories
}

// GetConfig 返回工厂的配置实例
// 返回值:
//   - *config.Config: 配置实例
func (f *Factory) GetConfig() *config.Config {
	return f.config
}

// ============================================================================
// 便捷分析函数
// ============================================================================

// analyzeFile 分析单个文件（私有函数）。
// 根据 input-parser 包的设计，仅支持 .sql 和 .log 文件类型。
// 参数:
//   - filePath: 文件路径，必须是有效的文件路径
//   - sqlParser: SQL 解析器实例，用于将 SQL 文本解析为 AST
//   - checkers: 检查器列表，用于检测兼容性问题和执行转换
//
// 返回:
//   - model.AnalysisResult: 分析结果，包含原始 SQL、发现的问题和转换后的 SQL
//   - error: 如果文件读取失败、创建分析器失败或 SQL 解析失败，返回错误
//
// 文件类型识别规则:
//   - .sql: 使用 SQL 文件解析器
//   - .log: 使用日志文件解析器
//   - 其他扩展名: 返回错误，不支持
func analyzeFile(filePath string, sqlParser sqlparser.SQLParser, checkers []checker.Checker) (model.AnalysisResult, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	inputParser, err := newFileParserForExt(ext)
	if err != nil {
		return model.AnalysisResult{Source: filePath}, err
	}

	// 使用对应的 inputParser 创建分析器
	analyzer, err := NewSQLAnalyzer(inputParser, sqlParser, checkers)
	if err != nil {
		return model.AnalysisResult{}, fmt.Errorf("创建分析器失败: %w", err)
	}

	// 直接调用 inputParser.Parse() 和 AnalyzeSQL()
	content, err := inputParser.Parse(filePath)
	if err != nil {
		return model.AnalysisResult{
			Source: filePath,
		}, fmt.Errorf("解析输入失败: %w", err)
	}

	result, err := analyzer.AnalyzeSQL(content, filePath)
	if result.Issues == nil {
		result.Issues = []model.Issue{}
	}
	return result, err
}

// analyzeDirectory 分析目录中的所有 SQL 相关文件（私有函数）。
// 递归遍历目录，分析所有 .sql、.log文件，并汇总所有发现的问题。
// 参数:
//   - dirPath: 目录路径，必须是有效的目录路径
//   - sqlParser: SQL 解析器实例，用于将 SQL 文本解析为 AST
//   - checkers: 检查器列表，用于检测兼容性问题和执行转换
//
// 返回:
//   - model.AnalysisResult: 分析结果，包含目录路径和所有文件的问题汇总
//   - error: 如果目录访问失败或遍历目录时出错，返回错误
//
// 注意:
//   - 单个文件分析失败不会中断整个目录遍历，错误会记录到 issues 中
//   - 支持的文件类型：.sql（SQL 文件）、.log（日志文件）
func analyzeDirectory(dirPath string, sqlParser sqlparser.SQLParser, checkers []checker.Checker) (model.AnalysisResult, error) {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return model.AnalysisResult{
			Source: dirPath,
		}, fmt.Errorf("访问目录失败: %w", err)
	}

	if !fileInfo.IsDir() {
		return model.AnalysisResult{
			Source: dirPath,
		}, fmt.Errorf("%s 不是目录", dirPath)
	}

	var allIssues []model.Issue

	// 遍历目录
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !isSupportedFileExt(ext) {
			return nil
		}

		// 分析文件，错误记录到 issues 但不中断遍历
		result, err := analyzeFile(path, sqlParser, checkers)
		if err != nil {
			// 记录错误但不中断目录遍历
			allIssues = append(allIssues, model.Issue{
				Checker: "Error",
				Message: fmt.Sprintf("分析文件失败: %v", err),
				Line:    0,
				File:    path,
			})
			return nil
		}

		allIssues = append(allIssues, result.Issues...)
		return nil
	})

	if err != nil {
		return model.AnalysisResult{
			Source: dirPath,
			Issues: allIssues,
		}, fmt.Errorf("遍历目录时出错: %w", err)
	}

	return model.AnalysisResult{
		Source: dirPath,
		Issues: allIssues,
	}, nil
}

// AnalyzeInput 分析输入源并返回结果。
// 这是分析功能的统一入口，支持多种输入类型：文件路径、目录路径、SQL 字符串、io.Reader。
// 函数会自动识别输入类型并调用相应的分析逻辑。
// 参数:
//   - source: 输入源，支持以下类型：
//   - string: 文件路径、目录路径或 SQL 字符串（自动识别）
//   - io.Reader: 流式输入，从 Reader 读取内容后分析
//   - sqlParser: SQL 解析器实例，用于将 SQL 文本解析为 AST
//   - checkers: 检查器列表，用于检测兼容性问题和执行转换
//
// 返回:
//   - model.AnalysisResult: 分析结果，包含原始 SQL、发现的问题和转换后的 SQL
//   - error: 如果输入类型不支持、文件/目录访问失败、创建分析器失败或 SQL 解析失败，返回错误
//
// 输入类型识别规则:
//   - string 类型：
//   - 如果路径存在且是目录：调用 analyzeDirectory
//   - 如果路径存在且是文件：调用 analyzeFile
//   - .log: 使用日志文件解析器
//   - 其他: 返回错误（不支持的文件类型）
//   - 如果路径不存在：作为 SQL 字符串处理
//   - io.Reader: 读取内容后作为 SQL 字符串分析
//
// 示例:
//
//	// 分析文件
//	result, err := AnalyzeInput("/path/to/file.sql", sqlParser, checkers)
//
//	// 分析目录
//	result, err := AnalyzeInput("/path/to/dir", sqlParser, checkers)
//
//	// 分析 SQL 字符串
//	result, err := AnalyzeInput("CREATE TABLE test (id INT)", sqlParser, checkers)
//
//	// 分析流输入
//	result, err := AnalyzeInput(strings.NewReader("SELECT * FROM users"), sqlParser, checkers)
func AnalyzeInput(source any, sqlParser sqlparser.SQLParser, checkers []checker.Checker) (model.AnalysisResult, error) {
	switch v := source.(type) {
	case string:
		// 检查是文件、目录还是SQL字符串
		fileInfo, err := os.Stat(v)
		if err == nil {
			// 路径存在，判断是目录还是文件
			if fileInfo.IsDir() {
				return analyzeDirectory(v, sqlParser, checkers)
			}

			// 是文件，直接使用 analyzeFile，它已经支持自动识别文件类型
			return analyzeFile(v, sqlParser, checkers)
		}

		// 路径不存在，作为 SQL 字符串处理
		analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
		if err != nil {
			return model.AnalysisResult{Source: v}, fmt.Errorf("创建分析器失败: %w", err)
		}
		return analyzer.AnalyzeSQL(v, "input_string")

	case io.Reader:
		// 读取内容并分析
		content, err := io.ReadAll(v)
		if err != nil {
			return model.AnalysisResult{
				Source: "io.Reader",
			}, fmt.Errorf("读取输入流失败: %w", err)
		}

		analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
		if err != nil {
			return model.AnalysisResult{Source: "io.Reader"}, fmt.Errorf("创建分析器失败: %w", err)
		}
		return analyzer.AnalyzeSQL(string(content), "io.Reader")

	default:
		return model.AnalysisResult{
			Source: "unknown",
		}, fmt.Errorf("不支持的输入类型: %T，仅支持 string 或 io.Reader", source)
	}
}
