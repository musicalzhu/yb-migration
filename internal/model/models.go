// Package model 提供分析结果与报告相关的数据模型。
// 包含 `Issue`, `AnalysisResult`, `Report` 等类型，用于分析和报告生成流程之间的数据传递。
package model

import (
	"time"
)

// AutoFix 表示自动修复的元数据
type AutoFix struct {
	Available bool   `json:"available"`
	Action    string `json:"action,omitempty"`
	Code      string `json:"code,omitempty"`
}

// Issue 表示兼容性问题的数据模型。
type Issue struct {
	Checker string  `json:"checker"`
	Message string  `json:"message"`
	File    string  `json:"file,omitempty"`
	Line    int     `json:"line,omitempty"`
	AutoFix AutoFix `json:"autofix,omitempty"`
}

// UniqueIssue 表示唯一的 issue 类型
type UniqueIssue struct {
	Checker string `json:"checker"` // 检查器名称
	Message string `json:"message"` // 问题描述
}

// AnalysisResult 表示 SQL 分析的结果
type AnalysisResult struct {
	SQL            string  `json:"sql"`                       // 原始 SQL 语句，可能包含多条 SQL 语句
	Issues         []Issue `json:"issues"`                    // 发现的问题列表
	Source         string  `json:"source,omitempty"`          // SQL 来源（文件、IO 等）
	TransformedSQL string  `json:"transformed_sql,omitempty"` // 转换后的SQL语句
}

// Report 表示 SQL 分析报告
// 针对 SQL 分片进行分析，不关心数据来源
type Report struct {
	TotalAnalyses int              `json:"total_analyses"` // 总分析项数量
	TotalIssues   int              `json:"total_issues"`   // 总问题数量（去重后）
	UniqueIssues  []UniqueIssue    `json:"unique_issues"`  // 唯一问题列表
	Results       []AnalysisResult `json:"results"`        // 每个分析项的结果
	GeneratedAt   time.Time        `json:"generated_at"`   // 报告生成时间

	// 新增字段：规则与检查器统计信息
	RuleStats    RuleStats    `json:"rule_stats"`    // 规则统计信息
	CheckerStats CheckerStats `json:"checker_stats"` // 检查器统计信息
}

// RuleStats 规则统计信息
type RuleStats struct {
	TotalRules int            `json:"total_rules"` // 总规则数量
	ByCategory map[string]int `json:"by_category"` // 按类别统计的规则数量
}

// CheckerStats 检查器统计信息
type CheckerStats struct {
	TotalCheckers int      `json:"total_checkers"` // 总检查器数量
	Checkers      []string `json:"checkers"`       // 检查器名称列表
}
