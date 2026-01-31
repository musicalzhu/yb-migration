package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/ybMigration/internal/checker"
	inputparser "github.com/example/ybMigration/internal/input-parser"
	sqlparser "github.com/example/ybMigration/internal/sql-parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestGenerateTransformedSQLFile 测试转换后SQL文件路径生成
func TestGenerateTransformedSQLFile(t *testing.T) {
	sourcePath := "/path/to/test.sql"
	outputDir := "/output/dir"

	expectedPath := filepath.Join(outputDir, "test_transformed.sql")
	actualPath := GenerateTransformedSQLFile(sourcePath, outputDir)

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
	err = SaveTransformedSQLToFile(result, outputPath)
	require.NoError(t, err)

	// 验证分析结果
	assert.Equal(t, sql, result.SQL)
	assert.Equal(t, "test", result.Source)
	assert.NotEmpty(t, result.TransformedSQL)

	// 验证文件存在且内容正确
	assert.FileExists(t, outputPath)

	content, err := os.ReadFile(outputPath)
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
		assert.Empty(t, result.Error)
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
		require.NoError(t, err)

		assert.Equal(t, sql, result.SQL)
		assert.Equal(t, "test", result.Source)
		assert.NotEmpty(t, result.Error)
		assert.Contains(t, result.Error, "未找到有效的 SQL 语句")
	})

	t.Run("invalid_sql", func(t *testing.T) {
		sql := "INVALID SQL STATEMENT"
		result, err := analyzer.AnalyzeSQL(sql, "test")
		require.NoError(t, err)

		assert.Equal(t, sql, result.SQL)
		assert.Equal(t, "test", result.Source)
		assert.NotEmpty(t, result.Error)
		assert.Contains(t, result.Error, "SQL 解析失败")
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

// TestAnalyzeFile 测试文件分析功能
func TestAnalyzeFile(t *testing.T) {
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
		result, err := AnalyzeFile(testDataPath, sqlParser, checkers)
		require.NoError(t, err)

		assert.NotEmpty(t, result.Source)
		assert.Empty(t, result.Error)
		// 文件包含JSON和GROUP_CONCAT，应该检测到问题
		assert.NotEmpty(t, result.Issues)
	})

	t.Run("analyze_nonexistent_file", func(t *testing.T) {
		_, err := AnalyzeFile("nonexistent.sql", sqlParser, checkers)
		require.Error(t, err) // AnalyzeFile对于不存在的文件确实返回错误

		// 当返回错误时，result是零值结构体，Error字段为空是正常的
		assert.Contains(t, err.Error(), "读取SQL文件失败")
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

		assert.NotEmpty(t, result.Error)
		assert.Contains(t, result.Error, "不支持的输入类型")
		assert.Contains(t, err.Error(), "不支持的输入类型")
	})
}
