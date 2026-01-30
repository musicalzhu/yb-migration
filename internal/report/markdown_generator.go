package report

import (
	"fmt"
	"os"

	"github.com/example/ybMigration/internal/model"
)

// MarkdownGenerator 生成 Markdown 格式的报告
type MarkdownGenerator struct{}

// Write 将分析结果写入 Markdown 文件
func (g *MarkdownGenerator) Write(path string, result model.AnalysisResult) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	// 写入报告标题
	fmt.Fprintln(f, "# SQL 分析报告")
	fmt.Fprintln(f)

	// 写入摘要信息
	fmt.Fprintln(f, "## 摘要")
	fmt.Fprintln(f)
	if result.Source != "" {
		fmt.Fprintf(f, "- **来源**: %s\n", result.Source)
	}
	if result.SQL != "" {
		fmt.Fprintln(f, "- **SQL 语句**:")
		fmt.Fprintln(f, "  ```sql")
		fmt.Fprintf(f, "  %s\n", result.SQL)
		fmt.Fprintln(f, "  ```")
	}
	if len(result.Issues) > 0 {
		fmt.Fprintf(f, "- **发现的问题数**: %d\n", len(result.Issues))
	} else {
		fmt.Fprintln(f, "- **状态**: 未发现兼容性问题")
	}
	fmt.Fprintln(f)

	// 写入问题详情
	if len(result.Issues) > 0 {
		fmt.Fprintln(f, "## 发现的问题")
		fmt.Fprintln(f)

		for i, issue := range result.Issues {
			title := fmt.Sprintf("问题 %d", i+1)
			if issue.Checker != "" {
				title = fmt.Sprintf("%s: %s", title, issue.Checker)
			}
			fmt.Fprintf(f, "### %s\n", title)
			fmt.Fprintln(f)
			fmt.Fprintf(f, "- **描述**: %s\n", issue.Message)
			if issue.File != "" {
				fmt.Fprintf(f, "- **文件**: %s", issue.File)
				if issue.Line > 0 {
					fmt.Fprintf(f, " (行号: %d)", issue.Line)
				}
				fmt.Fprintln(f)
			}

			if issue.AutoFix.Available && issue.AutoFix.Code != "" {
				fmt.Fprintln(f, "- **自动修复**: 可用")
				if issue.AutoFix.Action != "" {
					fmt.Fprintf(f, "  - **操作**: %s\n", issue.AutoFix.Action)
				}
				if issue.AutoFix.Code != "" {
					fmt.Fprintln(f, "  - **修复代码**:")
					fmt.Fprintln(f, "    ```sql")
					fmt.Fprintf(f, "    %s\n", issue.AutoFix.Code)
					fmt.Fprintln(f, "    ```")
				}
			} else {
				fmt.Fprintln(f, "- **自动修复**: 不可用")
			}
			fmt.Fprintln(f)
			fmt.Fprintln(f, "---")
			fmt.Fprintln(f)
		}
	}

	return nil
}
