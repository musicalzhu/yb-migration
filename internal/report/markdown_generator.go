package report

import (
	"bytes"
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
	if result.Source != "" {
		fmt.Fprintf(&buf, "- **来源**: %s\n", result.Source)
	}
	if result.SQL != "" {
		fmt.Fprintln(&buf, "- **SQL 语句**:")
		fmt.Fprintln(&buf, "  ```sql")
		fmt.Fprintf(&buf, "  %s\n", result.SQL)
		fmt.Fprintln(&buf, "  ```")
	}
	if len(result.Issues) > 0 {
		fmt.Fprintf(&buf, "- **发现的问题数**: %d\n", len(result.Issues))
	} else {
		fmt.Fprintln(&buf, "- **状态**: 未发现兼容性问题")
	}
	fmt.Fprintln(&buf)

	// 写入问题详情
	if len(result.Issues) > 0 {
		fmt.Fprintln(&buf, "## 发现的问题")
		fmt.Fprintln(&buf)

		for i, issue := range result.Issues {
			title := fmt.Sprintf("问题 %d", i+1)
			if issue.Checker != "" {
				title = fmt.Sprintf("%s: %s", title, issue.Checker)
			}
			fmt.Fprintf(&buf, "### %s\n", title)
			fmt.Fprintln(&buf)
			fmt.Fprintf(&buf, "- **描述**: %s\n", issue.Message)
			if issue.File != "" {
				fmt.Fprintf(&buf, "- **文件**: %s", issue.File)
				if issue.Line > 0 {
					fmt.Fprintf(&buf, " (行号: %d)", issue.Line)
				}
				fmt.Fprintln(&buf)
			}

			if issue.AutoFix.Available && issue.AutoFix.Code != "" {
				fmt.Fprintln(&buf, "- **自动修复**: 可用")
				if issue.AutoFix.Action != "" {
					fmt.Fprintf(&buf, "  - **操作**: %s\n", issue.AutoFix.Action)
				}
				if issue.AutoFix.Code != "" {
					fmt.Fprintln(&buf, "  - **修复代码**:")
					fmt.Fprintln(&buf, "    ```sql")
					fmt.Fprintf(&buf, "    %s\n", issue.AutoFix.Code)
					fmt.Fprintln(&buf, "    ```")
				}
			} else {
				fmt.Fprintln(&buf, "- **自动修复**: 不可用")
			}
			fmt.Fprintln(&buf)
			fmt.Fprintln(&buf, "---")
			fmt.Fprintln(&buf)
		}
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
