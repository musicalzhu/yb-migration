package analyzer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/example/ybMigration/internal/checker"
	inputparser "github.com/example/ybMigration/internal/input-parser"
	"github.com/example/ybMigration/internal/model"
	"github.com/example/ybMigration/internal/report"
	sqlparser "github.com/example/ybMigration/internal/sql-parser"
)

// TestNewSQLAnalyzer 测试SQL分析器创建
func TestNewSQLAnalyzer(t *testing.T) {
	sqlParser := sqlparser.NewSQLParser()
	checkers := []checker.Checker{}

	t.Run("valid_creation", func(t *testing.T) {
		analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
		require.NoError(t, err)
		assert.NotNil(t, analyzer)
	})

	t.Run("nil_parser", func(t *testing.T) {
		analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), nil, checkers)
		require.NoError(t, err) // NewSQLAnalyzer不检查nil parser，在运行时才会失败
		assert.NotNil(t, analyzer)
	})
}

// TestGenerateTransformedSQLPath 测试转换后SQL文件路径生成
func TestGenerateTransformedSQLPath(t *testing.T) {
	sourcePath := "/path/to/test.sql"
	outputDir := "/output/dir"

	expectedPath := filepath.Join(outputDir, "test_transformed.sql")
	actualPath := report.GenerateTransformedSQLPath(sourcePath, outputDir)

	assert.Equal(t, expectedPath, actualPath)
}

// TestAnalyzeSQLWithOutput 测试带输出的SQL分析
func TestAnalyzeSQLWithOutput(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "transformed.sql")

	// 创建分析器
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	sqlParser := sqlparser.NewSQLParser()
	checkers, err := factory.CreateCheckers("datatype")
	require.NoError(t, err)

	analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
	require.NoError(t, err)

	// 分析SQL（使用纯粹的AnalyzeSQL）
	sql := "CREATE TABLE test (id INT);"
	result, err := analyzer.AnalyzeSQL(sql, "test")
	require.NoError(t, err)

	// 手动保存转换后的SQL（使用新的简化API）
	err = report.SaveTransformedSQL(result, outputPath)
	require.NoError(t, err)

	// 验证分析结果
	assert.Equal(t, sql, result.SQL)
	assert.Equal(t, "test", result.Source)
	assert.NotEmpty(t, result.TransformedSQL)

	// 验证文件存在且内容正确
	assert.FileExists(t, outputPath)

	content, err := os.ReadFile(outputPath) //nolint:gosec
	require.NoError(t, err)
	assert.NotEmpty(t, content)

	t.Logf("原始SQL: %s", sql)
	t.Logf("转换后SQL: %s", string(content))
}

// TestSQLAnalyzer_AnalyzeSQL 测试SQL分析功能
func TestSQLAnalyzer_AnalyzeSQL(t *testing.T) {
	sqlParser := sqlparser.NewSQLParser()

	// 创建分析器工厂
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	// 创建检查器
	checkers, err := factory.CreateCheckers("datatype", "function")
	require.NoError(t, err)

	analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
	require.NoError(t, err)

	t.Run("valid_create_table", func(t *testing.T) {
		sql := "CREATE TABLE users (id INT, name VARCHAR(255))"
		result, err := analyzer.AnalyzeSQL(sql, "test")
		require.NoError(t, err)

		assert.Equal(t, sql, result.SQL)
		assert.Equal(t, "test", result.Source)
		assert.NotEmpty(t, result.TransformedSQL)
		// 移除 result.Error 检查，因为现在错误通过返回值处理
	})

	t.Run("sql_with_tinyint_datatype", func(t *testing.T) {
		sql := "CREATE TABLE users (id TINYINT, name VARCHAR(255))"
		result, err := analyzer.AnalyzeSQL(sql, "test")
		require.NoError(t, err)

		assert.Equal(t, sql, result.SQL)
		assert.Equal(t, "test", result.Source)
		assert.NotEmpty(t, result.TransformedSQL)

		// 应该检测到TINYINT数据类型兼容性问题
		assert.NotEmpty(t, result.Issues)
	})

	t.Run("sql_with_group_concat", func(t *testing.T) {
		sql := "SELECT GROUP_CONCAT(name) FROM users"
		result, err := analyzer.AnalyzeSQL(sql, "test")
		require.NoError(t, err)

		assert.Equal(t, sql, result.SQL)
		assert.Equal(t, "test", result.Source)
		assert.NotEmpty(t, result.TransformedSQL)

		// 应该检测到GROUP_CONCAT函数兼容性问题
		assert.NotEmpty(t, result.Issues)
	})

	t.Run("empty_sql", func(t *testing.T) {
		sql := ""
		result, err := analyzer.AnalyzeSQL(sql, "test")
		require.Error(t, err) // 现在应该返回错误

		// 检查错误类型
		var analysisErr *model.AnalysisError
		require.True(t, errors.As(err, &analysisErr))
		assert.Equal(t, model.ErrorTypeNoSQL, analysisErr.Type)
		assert.Contains(t, analysisErr.Message, "未找到有效的 SQL 语句")
		assert.Equal(t, sql, result.SQL)
		assert.Equal(t, "test", result.Source)
	})

	t.Run("invalid_sql", func(t *testing.T) {
		sql := "INVALID SQL STATEMENT"
		result, err := analyzer.AnalyzeSQL(sql, "test")
		require.Error(t, err) // 现在应该返回错误

		// 检查错误类型
		var analysisErr *model.AnalysisError
		require.True(t, errors.As(err, &analysisErr))
		assert.Equal(t, model.ErrorTypeParse, analysisErr.Type)
		assert.Contains(t, analysisErr.Message, "SQL 解析失败")
		assert.Equal(t, sql, result.SQL)
		assert.Equal(t, "test", result.Source)
	})
}

// TestNewAnalyzerFactory 测试分析器工厂创建
func TestNewAnalyzerFactory(t *testing.T) {
	t.Run("valid_creation", func(t *testing.T) {
		factory, err := NewAnalyzerFactory("")
		require.NoError(t, err)
		assert.NotNil(t, factory)
	})

	t.Run("nonexistent_config", func(t *testing.T) {
		factory, err := NewAnalyzerFactory("nonexistent.yaml")
		require.NoError(t, err) // 应该使用默认配置
		assert.NotNil(t, factory)
	})
}

// TestAnalyzerFactory_CreateCheckers 测试检查器创建
func TestAnalyzerFactory_CreateCheckers(t *testing.T) {
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	t.Run("create_datatype_checker", func(t *testing.T) {
		checkers, err := factory.CreateCheckers("datatype")
		require.NoError(t, err)
		assert.Len(t, checkers, 1)

		// 使用类型断言验证检查器类型
		_, ok := checkers[0].(*checker.DataTypeChecker)
		assert.True(t, ok, "应该创建DataTypeChecker")
	})

	t.Run("create_function_checker", func(t *testing.T) {
		checkers, err := factory.CreateCheckers("function")
		require.NoError(t, err)
		assert.Len(t, checkers, 1)

		// 使用类型断言验证检查器类型
		_, ok := checkers[0].(*checker.FunctionChecker)
		assert.True(t, ok, "应该创建FunctionChecker")
	})

	t.Run("create_multiple_checkers", func(t *testing.T) {
		checkers, err := factory.CreateCheckers("datatype", "function", "syntax", "charset")
		require.NoError(t, err)
		assert.Len(t, checkers, 4)
	})

	t.Run("create_no_checkers", func(t *testing.T) {
		checkers, err := factory.CreateCheckers()
		require.NoError(t, err)
		assert.Empty(t, checkers)
	})

	t.Run("unsupported_category", func(t *testing.T) {
		checkers, err := factory.CreateCheckers("unsupported")
		require.Error(t, err)
		assert.Nil(t, checkers)
		assert.Contains(t, err.Error(), "不支持的检查器类别")
	})
}

// TestAnalyzerFactory_CreateCheckersFromConfig 测试根据配置自动创建检查器
func TestAnalyzerFactory_CreateCheckersFromConfig(t *testing.T) {
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	t.Run("create_from_default_config", func(t *testing.T) {
		checkers, err := factory.CreateCheckersFromConfig()
		require.NoError(t, err)

		// 默认配置包含所有类别，应该创建4个检查器
		assert.Len(t, checkers, 4)

		// 验证检查器类型（顺序可能不同，用类型断言检查）
		var foundDatatype, foundFunction, foundSyntax, foundCharset bool
		for _, ch := range checkers {
			switch ch.(type) {
			case *checker.DataTypeChecker:
				foundDatatype = true
			case *checker.FunctionChecker:
				foundFunction = true
			case *checker.SyntaxChecker:
				foundSyntax = true
			case *checker.CharsetChecker:
				foundCharset = true
			}
		}
		assert.True(t, foundDatatype, "应该包含 DataTypeChecker")
		assert.True(t, foundFunction, "应该包含 FunctionChecker")
		assert.True(t, foundSyntax, "应该包含 SyntaxChecker")
		assert.True(t, foundCharset, "应该包含 CharsetChecker")
	})

	t.Run("extract_categories_from_config", func(t *testing.T) {
		categories := factory.extractCategoriesFromConfig()

		// 默认配置应该包含所有类别
		expectedCategories := map[string]bool{
			"datatype": false,
			"function": false,
			"syntax":   false,
			"charset":  false,
		}

		for _, cat := range categories {
			if _, exists := expectedCategories[cat]; exists {
				expectedCategories[cat] = true
			}
		}

		for cat, found := range expectedCategories {
			assert.True(t, found, "应该包含类别: %s", cat)
		}
	})
}

// TestAnalyzeInput_File 测试 AnalyzeInput 分析文件功能
func TestAnalyzeInput_File(t *testing.T) {
	sqlParser := sqlparser.NewSQLParser()

	// 创建分析器工厂
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	// 创建检查器 - 包含datatype和function来检测文件中的问题
	checkers, err := factory.CreateCheckers("datatype", "function")
	require.NoError(t, err)

	t.Run("analyze_sql_file", func(t *testing.T) {
		// 使用绝对路径
		testDataPath := "../../testdata/mysql_queries.sql"
		result, err := AnalyzeInput(testDataPath, sqlParser, checkers)
		require.NoError(t, err)

		assert.NotEmpty(t, result.Source)
		// 文件包含JSON和GROUP_CONCAT，应该检测到问题
		assert.NotEmpty(t, result.Issues)
	})

	t.Run("analyze_nonexistent_file", func(t *testing.T) {
		_, err := AnalyzeInput("nonexistent.sql", sqlParser, checkers)
		require.Error(t, err) // AnalyzeInput对于不存在的文件会作为SQL字符串处理，但会返回解析错误

		// 不存在的文件会被当作SQL字符串处理，所以错误信息可能是SQL解析失败
		assert.True(t, strings.Contains(err.Error(), "读取SQL文件失败") || strings.Contains(err.Error(), "SQL 解析失败"))
	})
}

// TestAnalyzeInput 测试通用输入分析功能
func TestAnalyzeInput(t *testing.T) {
	sqlParser := sqlparser.NewSQLParser()

	// 创建分析器工厂
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	// 创建检查器
	checkers, err := factory.CreateCheckers("datatype")
	require.NoError(t, err)

	t.Run("analyze_sql_string", func(t *testing.T) {
		sql := "CREATE TABLE test (id TINYINT, name VARCHAR(255))"
		result, err := AnalyzeInput(sql, sqlParser, checkers)
		require.NoError(t, err)

		assert.Equal(t, sql, result.SQL)
		assert.Equal(t, "input_string", result.Source)
		assert.NotEmpty(t, result.Issues) // TINYINT应该被检测到
	})

	t.Run("analyze_unsupported_type", func(t *testing.T) {
		result, err := AnalyzeInput(123, sqlParser, checkers)
		require.Error(t, err) // AnalyzeInput对于不支持类型确实返回错误

		assert.Equal(t, "unknown", result.Source)
		assert.Contains(t, err.Error(), "不支持的输入类型")
		// 移除 result.Error 检查，因为现在错误通过返回值处理
	})
}

func TestAnalyzeFile_UnsupportedFileType(t *testing.T) {
	sqlParser := sqlparser.NewSQLParser()
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	checkers, err := factory.CreateCheckers("datatype")
	require.NoError(t, err)

	t.Run("unsupported_file_extension", func(t *testing.T) {
		// 测试不支持的文件扩展名
		result, err := analyzeFile("test.xyz", sqlParser, checkers)
		require.Error(t, err)

		assert.Equal(t, "test.xyz", result.Source)
		assert.Contains(t, err.Error(), "不支持的文件类型: .xyz，仅支持 .sql 和 .log 文件")
	})
}

// ============================================================================
// 字符集前缀简单测试
// ============================================================================

// TestCharsetPrefix_Simple 简单测试字符集前缀处理
// 提供快速验证字符集前缀去除功能的基础测试
func TestCharsetPrefix_Simple(t *testing.T) {
	// 创建分析器工厂和检查器
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	checkers, err := factory.CreateCheckers("syntax")
	require.NoError(t, err)

	sqlParser := sqlparser.NewSQLParser()
	analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
	require.NoError(t, err)

	// 测试包含字符串的SQL
	inputSQL := "UPDATE users SET name = 'test' WHERE id = 1"
	result, err := analyzer.AnalyzeSQL(inputSQL, "test")
	require.NoError(t, err)
	require.NotEmpty(t, result.TransformedSQL)

	// 验证不包含字符集前缀
	charsetPrefixes := []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"}
	for _, prefix := range charsetPrefixes {
		require.NotContains(t, result.TransformedSQL, prefix,
			"转换后的SQL不应包含字符集前缀: %s", prefix)
	}

	t.Logf("输入SQL: %s", inputSQL)
	t.Logf("转换SQL: %s", result.TransformedSQL)
}

// ============================================================================
// SQL 转换质量测试
// ============================================================================

// TestSQLTransformQuality 测试SQL转换内容质量
// 验证转换后的SQL内容质量，包括关键字、标识符、字符串格式等
func TestSQLTransformQuality(t *testing.T) {
	tests := []struct {
		name          string
		inputSQL      string
		expectedSQL   string
		shouldContain []string // 必须包含的内容
		shouldNotHave []string // 不应包含的内容（字符集前缀等）
		description   string
	}{
		{
			name:          "字符串字面量质量",
			inputSQL:      "UPDATE users SET name = 'test' WHERE id = 1",
			expectedSQL:   "UPDATE `users` SET `name`='test' WHERE `id`=1",
			shouldContain: []string{"UPDATE", "SET", "WHERE", "`users`", "`name`", "'test'"},
			shouldNotHave: []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"},
			description:   "验证字符串字面量转换质量和字符集前缀去除",
		},
		{
			name:          "函数参数质量",
			inputSQL:      "SELECT COALESCE(orderid, 'N/A') FROM orders",
			expectedSQL:   "SELECT COALESCE(`orderid`, 'N/A') FROM `orders`",
			shouldContain: []string{"SELECT", "COALESCE", "FROM", "`orderid`", "`orders`", "'N/A'"},
			shouldNotHave: []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"},
			description:   "验证函数参数转换质量",
		},
		{
			name:          "INSERT语句质量",
			inputSQL:      "INSERT INTO users (name) VALUES ('张三')",
			expectedSQL:   "INSERT INTO `users` (`name`) VALUES ('张三')",
			shouldContain: []string{"INSERT", "INTO", "VALUES", "`users`", "`name`", "'张三'"},
			shouldNotHave: []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"},
			description:   "验证INSERT语句和中文字符处理质量",
		},
		{
			name:          "多条语句质量",
			inputSQL:      "UPDATE users SET name = 'test'; INSERT INTO logs (msg) VALUES ('info')",
			expectedSQL:   "UPDATE `users` SET `name`='test';\nINSERT INTO `logs` (`msg`) VALUES ('info')",
			shouldContain: []string{"UPDATE", "INSERT", "VALUES", "`users`", "`logs`", "'test'", "'info'"},
			shouldNotHave: []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"},
			description:   "验证多条SQL语句的转换质量",
		},
		{
			name:          "复杂查询质量",
			inputSQL:      "SELECT u.name, COUNT(*) FROM users u JOIN orders o ON u.id = o.user_id WHERE u.status = 'active' GROUP BY u.name",
			expectedSQL:   "SELECT `u`.`name`,COUNT(1) FROM `users` AS `u` JOIN `orders` AS `o` ON `u`.`id`=`o`.`user_id` WHERE `u`.`status`='active' GROUP BY `u`.`name`",
			shouldContain: []string{"SELECT", "FROM", "JOIN", "WHERE", "GROUP BY", "`users`", "`orders`", "'active'"},
			shouldNotHave: []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"},
			description:   "验证复杂JOIN查询的转换质量",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试场景: %s", tt.description)

			// 创建分析器工厂和检查器
			factory, err := NewAnalyzerFactory("")
			require.NoError(t, err)

			checkers, err := factory.CreateCheckers("syntax")
			require.NoError(t, err)

			sqlParser := sqlparser.NewSQLParser()
			analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
			require.NoError(t, err)

			// 分析SQL
			result, err := analyzer.AnalyzeSQL(tt.inputSQL, "test")
			require.NoError(t, err)
			require.NotEmpty(t, result.TransformedSQL, "转换后的SQL不应为空")

			// 验证转换后的SQL内容质量
			assert.True(t, flexibleSQLMatch(result.TransformedSQL, tt.expectedSQL),
				"转换后的SQL应与预期匹配（允许反引号差异）\n期望: %s\n实际: %s", tt.expectedSQL, result.TransformedSQL)

			// 验证包含必要的关键字和标识符
			for _, contain := range tt.shouldContain {
				assert.True(t, containsIdentifier(result.TransformedSQL, contain),
					"转换后的SQL应包含标识符: %s", contain)
			}

			// 验证不包含字符集前缀
			for _, prefix := range tt.shouldNotHave {
				assert.NotContains(t, result.TransformedSQL, prefix,
					"转换后的SQL不应包含字符集前缀: %s", prefix)
			}

			// 验证SQL格式正确性
			assertValidSQLFormat(t, result.TransformedSQL)

			t.Logf("输入SQL: %s", tt.inputSQL)
			t.Logf("转换SQL: %s", result.TransformedSQL)
		})
	}
}

// TestCharsetPrefixComprehensive 全面测试字符集前缀处理
// 验证各种字符集前缀都能被正确去除，确保SQL输出干净
func TestCharsetPrefixComprehensive(t *testing.T) {
	charsetPrefixes := []string{
		"_UTF8MB4", "_utf8mb4", "_UTF8", "_utf8",
		"_LATIN1", "_latin1", "_BINARY", "_binary",
	}

	tests := []struct {
		name        string
		inputSQL    string
		description string
	}{
		{
			name:        "简单字符串",
			inputSQL:    "SELECT * FROM users WHERE name = 'test'",
			description: "简单字符串字面量",
		},
		{
			name:        "中文字符串",
			inputSQL:    "INSERT INTO users (name) VALUES ('张三')",
			description: "包含中文字符的字符串",
		},
		{
			name:        "特殊字符",
			inputSQL:    "UPDATE config SET value = '特殊符号@#$%' WHERE id = 1",
			description: "包含特殊字符的字符串",
		},
		{
			name:        "函数参数",
			inputSQL:    "SELECT CONCAT(first_name, ' ', last_name) AS full_name FROM employees",
			description: "函数中的字符串参数",
		},
		{
			name:        "多个字符串",
			inputSQL:    "INSERT INTO logs (action, details) VALUES ('login', 'user logged in')",
			description: "INSERT语句中的多个字符串",
		},
		{
			name:        "复杂表达式",
			inputSQL:    "SELECT CASE WHEN status = 'active' THEN 'enabled' ELSE 'disabled' END FROM users",
			description: "CASE表达式中的字符串",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试场景: %s", tt.description)

			// 创建分析器
			factory, err := NewAnalyzerFactory("")
			require.NoError(t, err)

			checkers, err := factory.CreateCheckers("syntax")
			require.NoError(t, err)

			sqlParser := sqlparser.NewSQLParser()
			analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
			require.NoError(t, err)

			// 分析SQL
			result, err := analyzer.AnalyzeSQL(tt.inputSQL, "test")
			require.NoError(t, err)
			require.NotEmpty(t, result.TransformedSQL)

			// 验证不包含任何字符集前缀
			for _, prefix := range charsetPrefixes {
				assert.NotContains(t, result.TransformedSQL, prefix,
					"场景 '%s' 不应包含字符集前缀: %s", tt.name, prefix)
			}

			// 验证SQL格式正确性
			assertValidSQLFormat(t, result.TransformedSQL)

			t.Logf("输入: %s", tt.inputSQL)
			t.Logf("输出: %s", result.TransformedSQL)
		})
	}
}

// TestSQLFormatCorrectness 测试SQL格式正确性
// 验证转换后的SQL符合标准格式规范：关键字大写、标识符反引号、字符串单引号等
func TestSQLFormatCorrectness(t *testing.T) {
	tests := []struct {
		name         string
		inputSQL     string
		formatChecks []func(string) bool
		description  string
	}{
		{
			name:     "关键字大写",
			inputSQL: "select * from users where id = 1",
			formatChecks: []func(string) bool{
				hasUppercaseKeywords,
				// 移除反引号检查
				hasProperSpacing,
			},
			description: "验证关键字大写",
		},
		{
			name:     "标识符格式",
			inputSQL: "UPDATE users SET name = 'test' WHERE id = 1",
			formatChecks: []func(string) bool{
				// 移除反引号检查
				hasProperSpacing,
				noCharsetPrefixes,
			},
			description: "验证标识符格式",
		},
		{
			name:     "字符串单引号",
			inputSQL: "INSERT INTO logs (msg) VALUES ('info')",
			formatChecks: []func(string) bool{
				hasSingleQuoteStrings,
				noCharsetPrefixes,
				hasProperSpacing,
			},
			description: "验证字符串使用单引号",
		},
		{
			name:     "复杂语句格式",
			inputSQL: "SELECT u.name, COUNT(*) FROM users u JOIN orders o ON u.id = o.user_id WHERE u.status = 'active'",
			formatChecks: []func(string) bool{
				hasUppercaseKeywords,
				// 移除反引号检查
				hasProperSpacing,
				noCharsetPrefixes,
			},
			description: "验证复杂SQL语句格式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("测试场景: %s", tt.description)

			// 创建分析器
			factory, err := NewAnalyzerFactory("")
			require.NoError(t, err)

			checkers, err := factory.CreateCheckers("syntax")
			require.NoError(t, err)

			sqlParser := sqlparser.NewSQLParser()
			analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
			require.NoError(t, err)

			// 分析SQL
			result, err := analyzer.AnalyzeSQL(tt.inputSQL, "test")
			require.NoError(t, err)
			require.NotEmpty(t, result.TransformedSQL)

			// 执行格式检查
			for i, check := range tt.formatChecks {
				assert.True(t, check(result.TransformedSQL),
					"格式检查 %d 失败", i+1)
			}

			t.Logf("输入: %s", tt.inputSQL)
			t.Logf("输出: %s", result.TransformedSQL)
		})
	}
}

// TestFileOutputQuality 测试文件输出质量
// 验证转换后的SQL保存到文件时的内容一致性和格式正确性
func TestFileOutputQuality(t *testing.T) {
	// 创建临时目录
	tempDir := t.TempDir()

	// 测试SQL内容
	inputSQL := `UPDATE users SET name = 'test' WHERE id = 1;
INSERT INTO logs (action, details) VALUES ('login', 'user logged in');
SELECT COALESCE(orderid, 'N/A') FROM orders WHERE status = 'pending'`

	// 创建分析器
	factory, err := NewAnalyzerFactory("")
	require.NoError(t, err)

	checkers, err := factory.CreateCheckers("syntax")
	require.NoError(t, err)

	sqlParser := sqlparser.NewSQLParser()
	analyzer, err := NewSQLAnalyzer(inputparser.NewStringParser(), sqlParser, checkers)
	require.NoError(t, err)

	// 分析SQL
	result, err := analyzer.AnalyzeSQL(inputSQL, "test")
	require.NoError(t, err)
	require.NotEmpty(t, result.TransformedSQL)

	// 保存到文件
	outputPath := filepath.Join(tempDir, "transformed.sql")
	err = report.SaveTransformedSQL(result, outputPath)
	require.NoError(t, err)

	// 读取文件内容
	content, err := os.ReadFile(outputPath) //nolint:gosec
	require.NoError(t, err)

	savedSQL := string(content)

	// 验证文件内容质量
	assert.Equal(t, result.TransformedSQL, savedSQL, "文件内容应与转换结果一致")

	// 验证不包含字符集前缀
	charsetPrefixes := []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"}
	for _, prefix := range charsetPrefixes {
		assert.NotContains(t, savedSQL, prefix,
			"保存的SQL文件不应包含字符集前缀: %s", prefix)
	}

	// 验证SQL格式正确性
	assertValidSQLFormat(t, savedSQL)

	t.Logf("原始SQL: %s", inputSQL)
	t.Logf("转换SQL: %s", result.TransformedSQL)
	t.Logf("文件内容: %s", savedSQL)
}

// ============================================================================
// SQL 格式验证辅助函数
// ============================================================================

// assertValidSQLFormat 综合验证SQL格式正确性
func assertValidSQLFormat(t *testing.T, sql string) {
	assert.True(t, hasUppercaseKeywords(sql), "关键字应为大写")
	// 移除反引号检查，现在生成干净的SQL不使用反引号
	// assert.True(t, hasBacktickIdentifiers(sql), "标识符应使用反引号")
	assert.True(t, hasSingleQuoteStrings(sql), "字符串应使用单引号")
	assert.True(t, noCharsetPrefixes(sql), "不应包含字符集前缀")
	assert.True(t, hasProperSpacing(sql), "应有适当的空格")
}

// hasUppercaseKeywords 检查关键字是否为大写
func hasUppercaseKeywords(sql string) bool {
	keywords := []string{"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE", "JOIN", "GROUP", "ORDER", "BY"}
	for _, keyword := range keywords {
		if strings.Contains(sql, keyword) {
			return true
		}
	}
	return false
}

// containsIdentifier 灵活检查标识符，支持有或没有反引号
func containsIdentifier(sql, identifier string) bool {
	// 如果identifier本身包含反引号，先提取纯标识符名
	cleanIdentifier := strings.Trim(identifier, "`")

	// 检查无反引号的标识符
	if strings.Contains(sql, cleanIdentifier) {
		return true
	}

	// 检查有反引号的标识符
	backtickedIdentifier := fmt.Sprintf("`%s`", cleanIdentifier)
	if strings.Contains(sql, backtickedIdentifier) {
		return true
	}

	// 检查表别名形式（如 `u`.`name` 或 u.name）
	if strings.Contains(cleanIdentifier, ".") {
		parts := strings.Split(cleanIdentifier, ".")
		if len(parts) == 2 {
			// 检查无反引号的别名形式
			aliasedForm := fmt.Sprintf("%s.%s", parts[0], parts[1])
			if strings.Contains(sql, aliasedForm) {
				return true
			}

			// 检查有反引号的别名形式
			aliasedFormWithBackticks := fmt.Sprintf("`%s`.`%s`", parts[0], parts[1])
			if strings.Contains(sql, aliasedFormWithBackticks) {
				return true
			}
		}
	}

	return false
}

// flexibleSQLMatch 灵活的SQL匹配，允许有或没有反引号
func flexibleSQLMatch(actual, expected string) bool {
	// 如果完全匹配，直接返回true
	if actual == expected {
		return true
	}

	// 移除反引号后比较
	actualClean := strings.ReplaceAll(actual, "`", "")
	expectedClean := strings.ReplaceAll(expected, "`", "")

	return actualClean == expectedClean
}

// hasSingleQuoteStrings 检查字符串是否使用单引号
func hasSingleQuoteStrings(sql string) bool {
	return strings.Contains(sql, "'") && strings.Count(sql, "'")%2 == 0
}

// noCharsetPrefixes 检查是否没有字符集前缀
func noCharsetPrefixes(sql string) bool {
	charsetPrefixes := []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"}
	for _, prefix := range charsetPrefixes {
		if strings.Contains(sql, prefix) {
			return false
		}
	}
	return true
}

// hasProperSpacing 检查是否有适当的空格
func hasProperSpacing(sql string) bool {
	// 检查没有连续多个空格
	return !strings.Contains(sql, "  ") && !strings.Contains(sql, "\t")
}
