package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/ybMigration/internal/analyzer"
	report "github.com/example/ybMigration/internal/report"
	sqlparser "github.com/example/ybMigration/internal/sql-parser"
	"github.com/example/ybMigration/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 集成测试 - 完整工作流测试
// ============================================================================

// TestMain_Integration_ValidSQLFile 测试完整流程：配置加载 → SQL 文件解析 → 报告生成
func TestMain_Integration_ValidSQLFile(t *testing.T) {
	// 创建输出目录 - 放在 yb-migration/output-report/test 下
	reportPath, err := filepath.Abs("../output-report/test")
	require.NoError(t, err, "应能获取绝对路径")

	// 清理之前的输出
	os.RemoveAll(reportPath)

	// 确保目录存在
	err = os.MkdirAll(reportPath, 0755)
	require.NoError(t, err, "应能创建输出目录")

	// 使用 testutils 获取测试数据路径
	configPath := testutils.MustGetTestDataPath("../configs/default.yaml")
	sqlPath := testutils.MustGetTestDataPath("mysql_queries.sql")

	// 显式创建分析器和检查器并执行分析 + 报表生成（避免依赖 run 的内部行为）
	af, err := analyzer.NewAnalyzerFactory(configPath)
	require.NoError(t, err)
	checkers, err := af.CreateCheckers("function", "datatype", "syntax", "charset")
	require.NoError(t, err)
	sqlParser := sqlparser.NewSQLParser()
	require.NotNil(t, sqlParser)
	result, err := analyzer.AnalyzeInput(sqlPath, sqlParser, checkers)
	require.NoError(t, err)
	err = report.GenerateReports(reportPath, result)
	require.NoError(t, err)

	// 验证关键报告文件是否存在
	summaryPath := filepath.Join(reportPath, "summary.json")
	assert.FileExists(t, summaryPath, "应生成 summary.json 报告")

	// 验证报告内容
	data, err := os.ReadFile(summaryPath)
	require.NoError(t, err, "应能读取 summary.json")

	var summary map[string]interface{}
	err = json.Unmarshal(data, &summary)
	require.NoError(t, err, "summary.json 应为有效 JSON")

	// 验证报告结构
	assert.Contains(t, summary, "sql", "报告应包含 sql 字段")
	assert.Contains(t, summary, "issues", "报告应包含 issues 字段")
	assert.Contains(t, summary, "source", "报告应包含 source 字段")

	// 验证 issues 数组
	issues, ok := summary["issues"].([]interface{})
	require.True(t, ok, "issues 应为数组")
	assert.Greater(t, len(issues), 0, "应至少包含一个 issue")

	// 验证 issue 结构
	issue := issues[0].(map[string]interface{})
	assert.Contains(t, issue, "checker", "issue 应包含 checker 字段")
	assert.Contains(t, issue, "message", "issue 应包含 message 字段")

	t.Logf("报告生成在: %s", reportPath)
}

// TestMain_Integration_LogFile 测试日志文件分析
func TestMain_Integration_LogFile(t *testing.T) {
	// 创建输出目录 - 放在 yb-migration/output-report/log 下
	reportPath, err := filepath.Abs("../output-report/log")
	require.NoError(t, err, "应能获取绝对路径")

	// 清理之前的输出
	os.RemoveAll(reportPath)

	// 确保目录存在
	err = os.MkdirAll(reportPath, 0755)
	require.NoError(t, err, "应能创建输出目录")

	configPath := testutils.MustGetTestDataPath("../configs/default.yaml")
	logPath := testutils.MustGetTestDataPath("general_log_example.log")

	// 显式创建分析器和检查器并执行分析 + 报表生成
	af, err := analyzer.NewAnalyzerFactory(configPath)
	require.NoError(t, err)
	checkers, err := af.CreateCheckers("function", "datatype", "syntax", "charset")
	require.NoError(t, err)
	sqlParser := sqlparser.NewSQLParser()
	require.NotNil(t, sqlParser)
	result, err := analyzer.AnalyzeInput(logPath, sqlParser, checkers)
	require.NoError(t, err)
	err = report.GenerateReports(reportPath, result)
	require.NoError(t, err)

	// 验证报告文件生成
	summaryPath := filepath.Join(reportPath, "summary.json")
	assert.FileExists(t, summaryPath, "应生成 summary.json 报告")

	// 验证报告内容
	data, err := os.ReadFile(summaryPath)
	require.NoError(t, err, "应能读取 summary.json")

	var summary map[string]interface{}
	err = json.Unmarshal(data, &summary)
	require.NoError(t, err, "summary.json 应为有效 JSON")

	// 验证从日志中提取的 SQL
	issues, ok := summary["issues"].([]interface{})
	require.True(t, ok, "issues 应为数组")
	assert.GreaterOrEqual(t, len(issues), 1, "应至少包含一个 issue")

	t.Logf("日志分析报告生成在: %s", reportPath)
}

// TestMain_Integration_Directory 测试目录批量分析
func TestMain_Integration_Directory(t *testing.T) {
	// 创建输出目录 - 放在 yb-migration/output-report/directory 下
	reportPath, err := filepath.Abs("../output-report/directory")
	require.NoError(t, err, "应能获取绝对路径")

	// 清理之前的输出
	os.RemoveAll(reportPath)

	// 确保目录存在
	err = os.MkdirAll(reportPath, 0755)
	require.NoError(t, err, "应能创建输出目录")

	configPath := testutils.MustGetTestDataPath("../configs/default.yaml")
	testDir := testutils.MustGetTestDataPath("")

	// 显式创建分析器和检查器并执行分析 + 报表生成
	af, err := analyzer.NewAnalyzerFactory(configPath)
	require.NoError(t, err)
	checkers, err := af.CreateCheckers("function", "datatype", "syntax", "charset")
	require.NoError(t, err)
	sqlParser := sqlparser.NewSQLParser()
	require.NotNil(t, sqlParser)
	result, err := analyzer.AnalyzeInput(testDir, sqlParser, checkers)
	require.NoError(t, err)
	err = report.GenerateReports(reportPath, result)
	require.NoError(t, err)

	// 验证报告文件生成
	summaryPath := filepath.Join(reportPath, "summary.json")
	assert.FileExists(t, summaryPath, "应生成 summary.json 报告")

	// 验证报告内容
	data, err := os.ReadFile(summaryPath)
	require.NoError(t, err, "应能读取 summary.json")

	var summary map[string]interface{}
	err = json.Unmarshal(data, &summary)
	require.NoError(t, err, "summary.json 应为有效 JSON")

	// 验证从目录中分析到的问题
	issues, ok := summary["issues"].([]interface{})
	require.True(t, ok, "issues 应为数组")
	assert.GreaterOrEqual(t, len(issues), 2, "应至少包含两个 issue（SQL文件和日志文件）")

	t.Logf("目录分析报告生成在: %s", reportPath)
}

// TestMain_Integration_MultipleReportFormats 测试多种报告格式生成
func TestMain_Integration_MultipleReportFormats(t *testing.T) {
	// 创建输出目录 - 放在 yb-migration/output-report/formats 下
	reportPath, err := filepath.Abs("../output-report/formats")
	require.NoError(t, err, "应能获取绝对路径")

	// 清理之前的输出
	os.RemoveAll(reportPath)

	// 确保目录存在
	err = os.MkdirAll(reportPath, 0755)
	require.NoError(t, err, "应能创建输出目录")

	configPath := testutils.MustGetTestDataPath("../configs/default.yaml")
	sqlPath := testutils.MustGetTestDataPath("mysql_queries.sql")

	// 显式创建分析器和检查器并执行分析 + 报表生成
	af, err := analyzer.NewAnalyzerFactory(configPath)
	require.NoError(t, err)
	checkers, err := af.CreateCheckers("function", "datatype", "syntax", "charset")
	require.NoError(t, err)
	sqlParser := sqlparser.NewSQLParser()
	require.NotNil(t, sqlParser)
	result, err := analyzer.AnalyzeInput(sqlPath, sqlParser, checkers)
	require.NoError(t, err)
	err = report.GenerateReports(reportPath, result)
	require.NoError(t, err)

	// 验证主要报告文件生成
	summaryPath := filepath.Join(reportPath, "summary.json")
	assert.FileExists(t, summaryPath, "应生成 summary.json 报告")

	// 验证报告内容
	data, err := os.ReadFile(summaryPath)
	require.NoError(t, err, "应能读取 summary.json")

	var summary map[string]interface{}
	err = json.Unmarshal(data, &summary)
	require.NoError(t, err, "summary.json 应为有效 JSON")

	// 验证 HTML 报告（通常都会生成）
	htmlPath := filepath.Join(reportPath, "summary.html")
	if _, err := os.Stat(htmlPath); err == nil {
		t.Logf("HTML 报告已生成: %s", htmlPath)
	}

	// 列出所有生成的文件
	files, err := os.ReadDir(reportPath)
	require.NoError(t, err, "应能读取报告目录")

	t.Logf("生成的报告文件:")
	for _, file := range files {
		t.Logf("  - %s", file.Name())
	}
}

// TestMain_Integration_ErrorHandling 测试错误处理
func TestMain_Integration_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		inputPath   string
		reportPath  string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid_config_path",
			configPath:  testutils.MustGetTestDataPath("../nonexistent/config.yaml"),
			inputPath:   testutils.MustGetTestDataPath("mysql_queries.sql"),
			reportPath:  t.TempDir(),
			expectError: true,
			errorMsg:    "配置文件验证失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.configPath, tt.inputPath, tt.reportPath)

			if tt.expectError {
				assert.Error(t, err, "应该返回错误")
				assert.Contains(t, err.Error(), tt.errorMsg, "错误信息应包含预期内容")
			} else {
				assert.NoError(t, err, "不应该返回错误")
			}
		})
	}
}
