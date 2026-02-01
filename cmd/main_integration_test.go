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
	if err := os.RemoveAll(reportPath); err != nil {
		t.Logf("清理输出目录失败: %v", err)
	}

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
	err = report.GenerateReports(reportPath, result, af.GetConfig(), checkers)
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
	assert.Contains(t, summary, "results", "报告应包含 results 字段")
	assert.Contains(t, summary, "rule_stats", "报告应包含 rule_stats 字段")
	assert.Contains(t, summary, "checker_stats", "报告应包含 checker_stats 字段")

	// 验证 results 数组
	results, ok := summary["results"].([]interface{})
	require.True(t, ok, "results 应为数组")
	assert.Greater(t, len(results), 0, "应至少包含一个 result")

	// 验证第一个 result 结构
	firstResult := results[0].(map[string]interface{})
	assert.Contains(t, firstResult, "sql", "result 应包含 sql 字段")
	assert.Contains(t, firstResult, "issues", "result 应包含 issues 字段")
	assert.Contains(t, firstResult, "source", "result 应包含 source 字段")

	// 验证 issues 数组
	issues := firstResult["issues"].([]interface{})
	assert.Greater(t, len(issues), 0, "应至少包含一个 issue")

	// 验证 issue 结构
	issue := issues[0].(map[string]interface{})
	assert.Contains(t, issue, "checker", "issue 应包含 checker 字段")
	assert.Contains(t, issue, "message", "issue 应包含 message 字段")

	// 验证规则统计
	ruleStats := summary["rule_stats"].(map[string]interface{})
	assert.Contains(t, ruleStats, "total_rules", "应包含总规则数")
	assert.Contains(t, ruleStats, "by_category", "应包含按类别统计")

	// 验证检查器统计
	checkerStats := summary["checker_stats"].(map[string]interface{})
	assert.Contains(t, checkerStats, "total_checkers", "应包含总检查器数")
	assert.Contains(t, checkerStats, "checkers", "应包含检查器列表")

	t.Logf("报告生成在: %s", reportPath)
}

// TestMain_Integration_LogFile 测试日志文件分析
func TestMain_Integration_LogFile(t *testing.T) {
	// 创建输出目录 - 放在 yb-migration/output-report/log 下
	reportPath, err := filepath.Abs("../output-report/log")
	require.NoError(t, err, "应能获取绝对路径")

	// 清理之前的输出
	if err := os.RemoveAll(reportPath); err != nil {
		t.Logf("清理输出目录失败: %v", err)
	}

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
	err = report.GenerateReports(reportPath, result, af.GetConfig(), checkers)
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
	results := summary["results"].([]interface{})
	firstResult := results[0].(map[string]interface{})
	issues := firstResult["issues"].([]interface{})
	assert.GreaterOrEqual(t, len(issues), 1, "应至少包含一个 issue")

	// 验证转换后的SQL不包含字符集前缀
	transformedSQL := firstResult["transformed_sql"].(string)
	charsetPrefixes := []string{"_UTF8MB4", "_utf8", "_LATIN1", "_latin1", "_binary"}
	for _, prefix := range charsetPrefixes {
		assert.NotContains(t, transformedSQL, prefix,
			"转换后的SQL不应包含字符集前缀: %s", prefix)
	}

	t.Logf("日志分析报告生成在: %s", reportPath)
	t.Logf("转换后的SQL: %s", transformedSQL)
}

// TestMain_Integration_Directory 测试目录批量分析
func TestMain_Integration_Directory(t *testing.T) {
	// 创建输出目录 - 放在 yb-migration/output-report/directory 下
	reportPath, err := filepath.Abs("../output-report/directory")
	require.NoError(t, err, "应能获取绝对路径")

	// 清理之前的输出
	if err := os.RemoveAll(reportPath); err != nil {
		t.Logf("清理输出目录失败: %v", err)
	}

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
	err = report.GenerateReports(reportPath, result, af.GetConfig(), checkers)
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
	results := summary["results"].([]interface{})
	var allIssues []interface{}
	for _, result := range results {
		resultMap := result.(map[string]interface{})
		issues := resultMap["issues"].([]interface{})
		allIssues = append(allIssues, issues...)
	}
	assert.GreaterOrEqual(t, len(allIssues), 2, "应至少包含两个 issue（SQL文件和日志文件）")

	t.Logf("目录分析报告生成在: %s", reportPath)
}

// TestMain_Integration_MultipleReportFormats 测试多种报告格式生成
func TestMain_Integration_MultipleReportFormats(t *testing.T) {
	// 创建输出目录 - 放在 yb-migration/output-report/formats 下
	reportPath, err := filepath.Abs("../output-report/formats")
	require.NoError(t, err, "应能获取绝对路径")

	// 清理之前的输出
	if err := os.RemoveAll(reportPath); err != nil {
		t.Logf("清理输出目录失败: %v", err)
	}

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
	err = report.GenerateReports(reportPath, result, af.GetConfig(), checkers)
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
