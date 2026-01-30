package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	inputparser "github.com/example/ybMigration/internal/input-parser"
	sqlparser "github.com/example/ybMigration/internal/sql-parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
