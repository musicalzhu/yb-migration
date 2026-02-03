package report

import (
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/example/ybMigration/internal/model"
)

// HTMLGenerator 生成 HTML 格式的报告
type HTMLGenerator struct{}

// Write 将报告写入 HTML 文件
func (g *HTMLGenerator) Write(path string, report model.Report) error {
	// 准备模板数据
	data := struct {
		Title        string
		Report       model.Report
		DateTime     string
		TotalIssues  int
		RuleStats    model.RuleStats
		CheckerStats model.CheckerStats
	}{
		Title:        "SQL 分析报告",
		Report:       report,
		DateTime:     time.Now().Format("2006-01-02 15:04:05"),
		TotalIssues:  report.TotalIssues,
		RuleStats:    report.RuleStats,
		CheckerStats: report.CheckerStats,
	}

	// 解析模板
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	// 验证文件路径安全性
	if err := validateOutputPath(path); err != nil {
		return fmt.Errorf("不安全的文件路径: %w", err)
	}

	// 创建输出文件
	file, err := os.Create(path) //nolint:gosec
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "关闭文件 %s 失败: %v\n", path, err)
		}
	}()

	// 执行模板并写入文件
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("执行模板失败: %w", err)
	}

	return nil
}

// htmlTemplate 是 HTML 报告的模板
const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; margin: 0; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #333; border-bottom: 1px solid #eee; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        .summary { background: #f9f9f9; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .stats { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px; }
        .stat-box { background: #f8f9fa; padding: 15px; border-radius: 5px; border-left: 4px solid #007bff; }
        .issue { border: 1px solid #ddd; border-radius: 5px; padding: 15px; margin-bottom: 15px; }
        .success { color: #28a745; }
        .error { color: #dc3545; }
        .warning { color: #ffc107; }
        pre { background: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto; }
        .meta { color: #6c757d; font-size: 0.9em; }
        code { font-family: 'Courier New', Courier, monospace; }
        ul { list-style-type: none; padding-left: 0; }
        li { margin-bottom: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        
        <div class="summary">
            <h2>分析概览</h2>
            <p><strong>总分析项数:</strong> {{.Report.TotalAnalyses}}</p>
            <p><strong>总问题数:</strong> {{.TotalIssues}}</p>
            <p class="meta">生成时间: {{.DateTime}}</p>
        </div>

        <div class="stats">
            <div class="stat-box">
                <h3>规则统计</h3>
                <p><strong>总规则数:</strong> {{.RuleStats.TotalRules}}</p>
                {{if .RuleStats.ByCategory}}
                <h4>按类别统计:</h4>
                <ul>
                {{range $category, $count := .RuleStats.ByCategory}}
                    <li><strong>{{$category}}:</strong> {{$count}} 条规则</li>
                {{end}}
                </ul>
                {{end}}
            </div>
            
            <div class="stat-box">
                <h3>检查器统计</h3>
                <p><strong>总检查器数:</strong> {{.CheckerStats.TotalCheckers}}</p>
                {{if .CheckerStats.Checkers}}
                <h4>检查器列表:</h4>
                <ul>
                {{range $checker := .CheckerStats.Checkers}}
                    <li><strong>{{$checker}}</strong></li>
                {{end}}
                </ul>
                {{end}}
            </div>
        </div>

        {{if gt .TotalIssues 0}}
        <div class="issues">
            <h2>发现的问题 ({{.TotalIssues}})</h2>
            {{range $index, $issue := .Report.UniqueIssues}}
            <div class="issue">
                <h3>问题 #{{add $index 1}}: {{$issue.Checker}}</h3>
                <p>{{$issue.Message}}</p>
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="success-message">
            <p class="success">✓ 未发现兼容性问题</p>
        </div>
        {{end}}
    </div>
</body>
</html>`
