package report

import (
	"bytes"
	"fmt"
	"os"

	"github.com/example/ybMigration/internal/model"
)

// MarkdownGenerator 生成 Markdown 格式的报告
type MarkdownGenerator struct{}

// Write 将报告写入 Markdown 文件
func (g *MarkdownGenerator) Write(path string, report model.Report) error {
	// 验证文件路径安全性
	if err := validateOutputPath(path); err != nil {
		return fmt.Errorf("不安全的文件路径: %w", err)
	}

	f, err := os.Create(path) //nolint:gosec
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭文件 %s 失败: %v\n", path, err)
		}
	}()

	var buf bytes.Buffer

	// 写入报告标题
	fmt.Fprintln(&buf, "# SQL 分析报告")
	fmt.Fprintln(&buf)

	// 写入摘要信息
	fmt.Fprintln(&buf, "## 摘要")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "- **总分析项数**: %d\n", report.TotalAnalyses)
	fmt.Fprintf(&buf, "- **总问题数**: %d\n", report.TotalIssues)
	fmt.Fprintf(&buf, "- **报告生成时间**: %s\n", report.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintln(&buf)

	// 写入规则统计信息
	fmt.Fprintln(&buf, "## 规则统计")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "- **总规则数**: %d\n", report.RuleStats.TotalRules)
	if len(report.RuleStats.ByCategory) > 0 {
		fmt.Fprintln(&buf, "- **按类别统计**:")
		for category, count := range report.RuleStats.ByCategory {
			fmt.Fprintf(&buf, "  - **%s**: %d 条规则\n", category, count)
		}
	}
	fmt.Fprintln(&buf)

	// 写入检查器统计信息
	fmt.Fprintln(&buf, "## 检查器统计")
	fmt.Fprintln(&buf)
	fmt.Fprintf(&buf, "- **总检查器数**: %d\n", report.CheckerStats.TotalCheckers)
	if len(report.CheckerStats.Checkers) > 0 {
		fmt.Fprintln(&buf, "- **检查器列表**:")
		for _, checker := range report.CheckerStats.Checkers {
			fmt.Fprintf(&buf, "  - **%s**\n", checker)
		}
	}
	fmt.Fprintln(&buf)

	// 写入问题详情
	if len(report.UniqueIssues) > 0 {
		fmt.Fprintln(&buf, "## 发现的问题")
		fmt.Fprintln(&buf)

		for i, issue := range report.UniqueIssues {
			title := fmt.Sprintf("问题 %d", i+1)
			if issue.Checker != "" {
				title = fmt.Sprintf("%s: %s", title, issue.Checker)
			}
			fmt.Fprintf(&buf, "### %s\n", title)
			fmt.Fprintln(&buf)
			fmt.Fprintf(&buf, "- **描述**: %s\n", issue.Message)
			fmt.Fprintln(&buf)
			fmt.Fprintln(&buf, "---")
			fmt.Fprintln(&buf)
		}
	} else {
		fmt.Fprintln(&buf, "## 状态")
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "✅ 未发现兼容性问题")
		fmt.Fprintln(&buf)
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
