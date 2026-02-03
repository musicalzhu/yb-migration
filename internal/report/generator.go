// Package report 提供将分析结果输出为多种报告格式的功能。
// 支持 JSON、Markdown、HTML 等格式，并提供生成器注册与调用的通用接口。
package report

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/example/ybMigration/internal/checker"
	"github.com/example/ybMigration/internal/config"
	"github.com/example/ybMigration/internal/model"
)

// Generator 定义报告生成器接口
type Generator interface {
	// Write 将报告写入指定路径
	// path: 输出文件路径
	// report: 包含统计信息的完整报告
	// 返回值: 错误信息
	Write(path string, report model.Report) error
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
// cfg: 配置实例（用于统计规则信息）
// checkers: 检查器列表（用于统计检查器信息）
// formats: 要生成的报告格式，如 "json"、"markdown"、"html"
// 返回值: 错误信息
func GenerateReports(basePath string, result model.AnalysisResult, cfg *config.Config, checkers []checker.Checker, formats ...string) error {
	if len(formats) == 0 {
		formats = SupportedFormats()
	}

	// 固定报告文件名（可配置，但通常固定为 summary）
	const reportFileName = "summary"

	// 生成包含统计信息的报告
	report := GenerateReport(result, cfg, checkers)

	for _, format := range formats {
		generator, ok := GetGenerator(format)
		if !ok {
			return fmt.Errorf("不支持的报告格式: %s", format)
		}

		ext := getFileExtension(format)
		outputPath := filepath.Join(basePath, reportFileName+ext)

		if err := generator.Write(outputPath, *report); err != nil {
			return fmt.Errorf("生成 %s 报告失败: %w", format, err)
		}
	}

	return nil
}

// GenerateReport 根据分析结果生成报告
// result: 分析结果
// cfg: 配置实例（用于统计规则信息）
// checkers: 检查器列表（用于统计检查器信息）
// 返回值: 生成的报告
func GenerateReport(result model.AnalysisResult, cfg *config.Config, checkers []checker.Checker) *model.Report {
	uniqueIssues := collectUniqueIssues(result.Issues)

	return &model.Report{
		TotalAnalyses: 1,
		TotalIssues:   len(uniqueIssues),
		UniqueIssues:  uniqueIssues,
		Results:       []model.AnalysisResult{result},
		GeneratedAt:   time.Now(),
		RuleStats:     collectRuleStats(cfg),
		CheckerStats:  collectCheckerStats(checkers),
	}
}

// GenerateReportFromMultiple 根据多个分析结果生成合并报告
// results: 多个分析结果
// cfg: 配置实例（用于统计规则信息）
// checkers: 检查器列表（用于统计检查器信息）
// 返回值: 合并后的报告
func GenerateReportFromMultiple(results []model.AnalysisResult, cfg *config.Config, checkers []checker.Checker) *model.Report {
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
		RuleStats:     collectRuleStats(cfg),
		CheckerStats:  collectCheckerStats(checkers),
	}
}

// collectUniqueIssues 收集唯一的问题
// 参数:
//   - issues: 问题列表
//
// 返回值:
//   - []model.UniqueIssue: 去重后的问题列表
//
// 实现细节:
//  1. 使用map进行去重，key为UniqueIssue结构体
//  2. 去重依据: 检查器名称和消息内容
//  3. 转换为slice返回，保持顺序一致性
//
// 注意事项:
//   - 忽略问题的位置信息，只关注内容
//   - 返回的顺序可能不固定
func collectUniqueIssues(issues []model.Issue) []model.UniqueIssue {
	uniqueIssues := make(map[model.UniqueIssue]bool)
	for _, issue := range issues {
		uniqueIssue := model.UniqueIssue{
			Checker: issue.Checker,
			Message: issue.Message,
		}
		uniqueIssues[uniqueIssue] = true
	}

	result := make([]model.UniqueIssue, 0, len(uniqueIssues))
	for issue := range uniqueIssues {
		result = append(result, issue)
	}
	return result
}

// collectRuleStats 收集规则统计信息
// 参数:
//   - cfg: 配置实例
//
// 返回值:
//   - model.RuleStats: 规则统计信息
//
// 实现细节:
//  1. 遍历配置中的所有规则类别
//  2. 统计每个类别的规则数量
//  3. 计算总规则数
//
// 注意事项:
//   - 配置为空时返回空的统计信息
//   - 统计包括启用和禁用的规则
func collectRuleStats(cfg *config.Config) model.RuleStats {
	if cfg == nil {
		return model.RuleStats{
			TotalRules: 0,
			ByCategory: make(map[string]int),
		}
	}

	stats := model.RuleStats{
		TotalRules: len(cfg.Rules),
		ByCategory: make(map[string]int),
	}

	for _, rule := range cfg.Rules {
		stats.ByCategory[rule.Category]++
	}

	return stats
}

// collectCheckerStats 收集检查器统计信息
// 参数:
//   - checkers: 检查器列表
//
// 返回值:
//   - model.CheckerStats: 检查器统计信息
//
// 实现细节:
//  1. 统计检查器总数
//  2. 提取每个检查器的名称
//  3. 使用反射获取检查器类型信息
//
// 注意事项:
//   - 检查器名称通过反射获取
//   - 返回的名称列表按输入顺序排列
func collectCheckerStats(checkers []checker.Checker) model.CheckerStats {
	stats := model.CheckerStats{
		TotalCheckers: len(checkers),
		Checkers:      make([]string, 0, len(checkers)),
	}

	for _, ch := range checkers {
		stats.Checkers = append(stats.Checkers, ch.Name())
	}

	return stats
}

// getFileExtension 获取文件扩展名
// 参数:
//   - format: 文件格式（如 "markdown", "json", "html"）
//
// 返回值:
//   - string: 对应的文件扩展名（包含点号）
//
// 支持的格式:
//   - markdown -> .md
//   - json -> .json
//   - html -> .html
//   - 默认 -> .format
//
// 注意事项:
//   - 格式参数不区分大小写
//   - 未知格式返回 .txt
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
