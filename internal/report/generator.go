// Package report 提供将分析结果输出为多种报告格式的功能。
// 支持 JSON、Markdown、HTML 等格式，并提供生成器注册与调用的通用接口。
package report

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/example/ybMigration/internal/model"
)

// Generator 定义报告生成器接口
type Generator interface {
	// Write 将分析结果写入指定路径
	// path: 输出文件路径
	// result: 分析结果
	// 返回值: 错误信息
	Write(path string, result model.AnalysisResult) error
}

// generatorRegistry 存储所有可用的报告生成器
var generatorRegistry = map[string]Generator{
	"json":     &JSONGenerator{},
	"markdown": &MarkdownGenerator{},
	"html":     &HTMLGenerator{},
}

// GetGenerator 获取指定格式的报告生成器
func GetGenerator(format string) (Generator, bool) {
	g, ok := generatorRegistry[format]
	return g, ok
}

// SupportedFormats 返回支持的报告格式列表
func SupportedFormats() []string {
	formats := make([]string, 0, len(generatorRegistry))
	for format := range generatorRegistry {
		formats = append(formats, format)
	}
	return formats
}

// GenerateReports 生成多种格式的报告
// basePath: 报告文件的基础路径（不包含扩展名）
// result: 分析结果
// formats: 要生成的报告格式，如 "json"、"markdown"、"html"
// 返回值: 错误信息
func GenerateReports(basePath string, result model.AnalysisResult, formats ...string) error {
	if len(formats) == 0 {
		formats = SupportedFormats()
	}

	// 固定报告文件名（可配置，但通常固定为 summary）
	const reportFileName = "summary"

	for _, format := range formats {
		generator, ok := GetGenerator(format)
		if !ok {
			return fmt.Errorf("不支持的报告格式: %s", format)
		}

		ext := getFileExtension(format)
		outputPath := filepath.Join(basePath, reportFileName+ext)

		if err := generator.Write(outputPath, result); err != nil {
			return fmt.Errorf("生成 %s 报告失败: %w", format, err)
		}
	}

	return nil
}

// GenerateReport 根据分析结果生成报告
// result: 分析结果
// 返回值: 生成的报告
func GenerateReport(result model.AnalysisResult) *model.Report {
	uniqueIssues := collectUniqueIssues(result.Issues)

	return &model.Report{
		TotalAnalyses: 1,
		TotalIssues:   len(uniqueIssues),
		UniqueIssues:  uniqueIssues,
		Results:       []model.AnalysisResult{result},
		GeneratedAt:   time.Now(),
	}
}

// GenerateReportFromMultiple 根据多个分析结果生成合并报告
// results: 多个分析结果
// 返回值: 合并后的报告
func GenerateReportFromMultiple(results []model.AnalysisResult) *model.Report {
	var allIssues []model.Issue
	for _, result := range results {
		allIssues = append(allIssues, result.Issues...)
	}

	uniqueIssues := collectUniqueIssues(allIssues)

	return &model.Report{
		TotalAnalyses: len(results),
		TotalIssues:   len(uniqueIssues),
		UniqueIssues:  uniqueIssues,
		Results:       results,
		GeneratedAt:   time.Now(),
	}
}

// collectUniqueIssues 收集唯一的问题
// issues: 问题列表
// 返回值: 去重后的问题列表
func collectUniqueIssues(issues []model.Issue) []model.UniqueIssue {
	uniqueIssues := make(map[model.UniqueIssue]bool)
	for _, issue := range issues {
		uniqueIssue := model.UniqueIssue{
			Checker: issue.Checker,
			Message: issue.Message,
		}
		uniqueIssues[uniqueIssue] = true
	}

	var result []model.UniqueIssue
	for issue := range uniqueIssues {
		result = append(result, issue)
	}
	return result
}

// getFileExtension 获取文件扩展名
// format: 文件格式
// 返回值: 对应的文件扩展名
func getFileExtension(format string) string {
	switch format {
	case "markdown":
		return ".md"
	case "json":
		return ".json"
	case "html":
		return ".html"
	default:
		return "." + format
	}
}
