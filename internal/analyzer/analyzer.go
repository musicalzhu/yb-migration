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
	sqlParser   sqlparser.SqlParser
	checkers    []checker.Checker
}

// NewSQLAnalyzer 创建 SQL 分析器
func NewSQLAnalyzer(inputParser inputparser.InputParser, sqlParser sqlparser.SqlParser, checkers []checker.Checker) (*SQLAnalyzer, error) {
	// 支持空的 checkers 列表（表示无需执行任何检查，完全兼容）
	// 直接返回 analyzer，不将其视为错误
	return &SQLAnalyzer{
		inputParser: inputParser,
		sqlParser:   sqlParser,
		checkers:    checkers,
	}, nil
}

// AnalyzeSQL 分析 SQL 语句（支持转换）
func (a *SQLAnalyzer) AnalyzeSQL(sql string, source string) (model.AnalysisResult, error) {
	// 解析 SQL 语句
	stmts, err := a.sqlParser.ParseSQL(sql)
	if err != nil {
		return model.AnalysisResult{
			SQL:    sql,
			Source: source,
			Error:  fmt.Sprintf("SQL 解析失败: %v", err),
		}, nil
	}

	if len(stmts) == 0 {
		return model.AnalysisResult{
			SQL:    sql,
			Source: source,
			Error:  "未找到有效的 SQL 语句",
		}, nil
	}

	// 使用 checker.Check 进行一次遍历完成分析和转换
	result := checker.Check(stmts, a.checkers...)

	// 生成转换后的SQL
	transformedSQL, err := a.generateSQL(result.TransformedStmts)
	if err != nil {
		return model.AnalysisResult{
			SQL:    sql,
			Source: source,
			Error:  fmt.Sprintf("生成转换SQL失败: %v", err),
		}, nil
	}

	return model.AnalysisResult{
		SQL:            sql,
		Source:         source,
		Issues:         result.Issues,
		TransformedSQL: transformedSQL,
	}, nil
}

// generateSQL 从AST节点生成SQL字符串（使用TiDB的Restore功能）
func (a *SQLAnalyzer) generateSQL(stmts []ast.StmtNode) (string, error) {
	if len(stmts) == 0 {
		return "", nil
	}

	// 使用TiDB的Restore API生成SQL
	var sqlParts []string
	for _, stmt := range stmts {
		if stmt != nil {
			// 创建strings.Builder作为RestoreWriter
			var builder strings.Builder

			// 创建RestoreCtx，使用默认标志
			ctx := format.NewRestoreCtx(format.DefaultRestoreFlags, &builder)

			// 调用TiDB的Restore方法生成SQL
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

	return strings.Join(sqlParts, ";\n"), nil
}

// ============================================================================
// 便捷函数
// ============================================================================

// GenerateTransformedSQLFile 生成转换后的SQL文件路径
func GenerateTransformedSQLFile(sourcePath, outputDir string) string {
	// 获取文件名（不含扩展名）
	baseName := filepath.Base(sourcePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)

	// 生成转换后的文件名
	transformedName := fmt.Sprintf("%s_transformed.sql", nameWithoutExt)
	return filepath.Join(outputDir, transformedName)
}

// SaveTransformedSQLToFile 保存转换后的SQL到文件（包级别函数）
func SaveTransformedSQLToFile(result model.AnalysisResult, outputPath string) error {
	if result.TransformedSQL == "" {
		return fmt.Errorf("没有转换后的SQL需要保存")
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 写入转换后的SQL文件
	if err := os.WriteFile(outputPath, []byte(result.TransformedSQL), 0644); err != nil {
		return fmt.Errorf("写入转换后的SQL文件失败: %w", err)
	}

	return nil
}

// ============================================================================
// 分析器工厂
// ============================================================================

// AnalyzerFactory 分析器工厂
type AnalyzerFactory struct {
	config *config.Config
}

// NewAnalyzerFactory 创建分析器工厂
func NewAnalyzerFactory(configPath string) (*AnalyzerFactory, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败: %w", err)
	}

	return &AnalyzerFactory{
		config: cfg,
	}, nil
}

// CreateCheckers 创建检查器列表
func (f *AnalyzerFactory) CreateCheckers(categories ...string) ([]checker.Checker, error) {
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

// CreateStringAnalyzer 创建字符串分析器
func (f *AnalyzerFactory) CreateStringAnalyzer(categories ...string) (*SQLAnalyzer, error) {
	checkers, err := f.CreateCheckers(categories...)
	if err != nil {
		return nil, err
	}

	return NewSQLAnalyzer(inputparser.NewStringParser(), sqlparser.NewSQLParser(), checkers)
}

// CreateFileAnalyzer 创建文件分析器
func (f *AnalyzerFactory) CreateFileAnalyzer(categories ...string) (*SQLAnalyzer, error) {
	checkers, err := f.CreateCheckers(categories...)
	if err != nil {
		return nil, err
	}

	return NewSQLAnalyzer(inputparser.NewSQLFileParser(), sqlparser.NewSQLParser(), checkers)
}

// ============================================================================
// 默认工厂实例
// ============================================================================

// NewDefaultAnalyzerFactory 创建默认分析器工厂
func NewDefaultAnalyzerFactory() (*AnalyzerFactory, error) {
	return NewAnalyzerFactory("")
}

// NewDefaultStringAnalyzer 创建默认字符串分析器
func NewDefaultStringAnalyzer() (*SQLAnalyzer, error) {
	factory, err := NewDefaultAnalyzerFactory()
	if err != nil {
		return nil, err
	}

	return factory.CreateStringAnalyzer()
}

// NewDefaultFileAnalyzer 创建默认文件分析器
func NewDefaultFileAnalyzer() (*SQLAnalyzer, error) {
	factory, err := NewDefaultAnalyzerFactory()
	if err != nil {
		return nil, err
	}

	return factory.CreateFileAnalyzer()
}

// ============================================================================
// 便捷分析函数
// ============================================================================

// AnalyzeFile 分析文件
func AnalyzeFile(filePath string, sqlParser sqlparser.SqlParser, checkers []checker.Checker) (model.AnalysisResult, error) {
	// 读取 SQL 文件内容
	fileParser := inputparser.NewSQLFileParser()
	content, err := fileParser.Parse(filePath)
	if err != nil {
		return model.AnalysisResult{}, fmt.Errorf("读取SQL文件失败: %w", err)
	}

	// 使用字符串解析器对文件内容进行分析
	analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
	if err != nil {
		return model.AnalysisResult{}, fmt.Errorf("创建分析器失败: %w", err)
	}

	result, err := analyzer.AnalyzeSQL(content, filePath)
	if result.Issues == nil {
		result.Issues = []model.Issue{}
	}
	return result, err
}

// AnalyzeDirectory 分析目录
func AnalyzeDirectory(dirPath string, sqlParser sqlparser.SqlParser, checkers []checker.Checker) (model.AnalysisResult, error) {
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return model.AnalysisResult{
			Source: dirPath,
			Error:  fmt.Sprintf("访问目录失败: %v", err),
		}, err
	}

	if !fileInfo.IsDir() {
		return model.AnalysisResult{
			Source: dirPath,
			Error:  fmt.Sprintf("%s 不是目录", dirPath),
		}, nil
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
		// 根据文件类型选择解析器：优先按扩展名分派
		var result model.AnalysisResult

		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".sql":
			result, err = AnalyzeFile(path, sqlParser, checkers)
			if err != nil {
				return err
			}
		case ".log", ".txt":
			// 使用 general log 解析器提取 SQL 并分析
			gl := inputparser.NewGeneralLogFileParser()
			sqlContent, perr := gl.ParseFile(path)
			if perr != nil {
				// 尝试按 SQL 文件解析作为回退
				result, err = AnalyzeFile(path, sqlParser, checkers)
				if err != nil {
					return err
				}
				break
			}
			analyzer, nerr := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
			if nerr != nil {
				return nerr
			}
			result, err = analyzer.AnalyzeSQL(sqlContent, path)
			if err != nil {
				return err
			}
		default:
			// 未知扩展名：先尝试按 SQL 文件解析，失败再尝试日志解析
			result, err = AnalyzeFile(path, sqlParser, checkers)
			if err != nil {
				gl := inputparser.NewGeneralLogFileParser()
				sqlContent, perr := gl.ParseFile(path)
				if perr != nil {
					return err
				}
				analyzer, nerr := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
				if nerr != nil {
					return nerr
				}
				result, err = analyzer.AnalyzeSQL(sqlContent, path)
				if err != nil {
					return err
				}
			}
		}

		if result.Error != "" {
			allIssues = append(allIssues, model.Issue{
				Checker: "Error",
				Message: result.Error,
				Line:    0,
				File:    path,
			})
		} else {
			if result.Issues == nil {
				// 保证 Issues 非 nil
				result.Issues = []model.Issue{}
			}
			allIssues = append(allIssues, result.Issues...)
		}

		return nil
	})

	if err != nil {
		return model.AnalysisResult{
			Source: dirPath,
			Error:  fmt.Sprintf("遍历目录时出错: %v", err),
		}, nil
	}

	return model.AnalysisResult{
		Source: dirPath,
		Issues: allIssues,
	}, nil
}

// AnalyzeInput 分析输入源并返回结果
func AnalyzeInput(source any, sqlParser sqlparser.SqlParser, checkers []checker.Checker) (model.AnalysisResult, error) {
	switch v := source.(type) {
	case string:
		// 检查是文件、目录还是SQL字符串
		fileInfo, err := os.Stat(v)
		if err == nil {
			if fileInfo.IsDir() {
				return AnalyzeDirectory(v, sqlParser, checkers)
			}

			// 根据扩展名分派文件解析：.sql 使用 AnalyzeFile，.log/.txt 使用 general log 解析
			ext := strings.ToLower(filepath.Ext(v))
			if ext == ".sql" {
				return AnalyzeFile(v, sqlParser, checkers)
			}

			if ext == ".log" || ext == ".txt" || strings.Contains(strings.ToLower(v), "log") {
				gl := inputparser.NewGeneralLogFileParser()
				sqlContent, perr := gl.ParseFile(v)
				if perr != nil {
					return model.AnalysisResult{Source: v, Error: fmt.Sprintf("读取日志文件失败: %v", perr)}, perr
				}
				analyzer, nerr := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
				if nerr != nil {
					return model.AnalysisResult{Source: v, Error: fmt.Sprintf("创建分析器失败: %v", nerr)}, nerr
				}
				return analyzer.AnalyzeSQL(sqlContent, v)
			}

			// 其他扩展名：默认尝试按 SQL 文件解析
			return AnalyzeFile(v, sqlParser, checkers)
		}

		// SQL字符串
		analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
		if err != nil {
			return model.AnalysisResult{
				Source: v,
				Error:  fmt.Sprintf("创建分析器失败: %v", err),
			}, err
		}
		return analyzer.AnalyzeSQL(v, "input_string")

	case io.Reader:
		// 读取内容并分析
		content, err := io.ReadAll(v)
		if err != nil {
			return model.AnalysisResult{
				Source: "io.Reader",
				Error:  fmt.Sprintf("读取输入流失败: %v", err),
			}, err
		}

		analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
		if err != nil {
			return model.AnalysisResult{
				Source: "io.Reader",
				Error:  fmt.Sprintf("创建分析器失败: %v", err),
			}, err
		}
		return analyzer.AnalyzeSQL(string(content), "io.Reader")

	default:
		return model.AnalysisResult{
			Source: "unknown",
			Error:  fmt.Sprintf("不支持的输入类型: %T", source),
		}, fmt.Errorf("不支持的输入类型: %T", source)
	}
}
