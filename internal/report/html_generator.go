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

// Write 将分析结果写入 HTML 文件
func (g *HTMLGenerator) Write(path string, result model.AnalysisResult) error {
	// 准备模板数据
	data := struct {
		Title       string
		Result      model.AnalysisResult
		DateTime    string
		TotalIssues int
	}{
		Title:       "SQL 分析报告",
		Result:      result,
		DateTime:    time.Now().Format("2006-01-02 15:04:05"),
		TotalIssues: len(result.Issues),
	}

	// 解析模板
	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %w", err)
	}

	// 创建输出文件
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

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
        .summary { background: #f9f9f9; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .issue { border: 1px solid #ddd; border-radius: 5px; padding: 15px; margin-bottom: 15px; }
        .success { color: #28a745; }
        .error { color: #dc3545; }
        .warning { color: #ffc107; }
        pre { background: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto; }
        .meta { color: #6c757d; font-size: 0.9em; }
        code { font-family: 'Courier New', Courier, monospace; }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        
        <div class="summary">
            <h2>分析概览</h2>
            {{if .Result.Source}}<p><strong>来源:</strong> {{.Result.Source}}</p>{{end}}
            <p><strong>问题数量:</strong> {{.TotalIssues}}</p>
            <p class="meta">生成时间: {{.DateTime}}</p>
        </div>

        {{if gt .TotalIssues 0}}
        <div class="issues">
            <h2>发现的问题 ({{.TotalIssues}})</h2>
            {{range $index, $issue := .Result.Issues}}
            <div class="issue">
                <h3>问题 #{{add $index 1}}: {{$issue.Checker}}</h3>
                <p>{{$issue.Message}}</p>
                {{if or $issue.File $issue.Line}}
                <p class="meta">
                    {{if $issue.File}}文件: {{$issue.File}}{{end}}
                    {{if $issue.Line}}, 行号: {{$issue.Line}}{{end}}
                </p>
                {{end}}
                {{if $issue.AutoFix.Available}}
                <p><strong>自动修复:</strong> 可用
                    {{if $issue.AutoFix.Action}}
                    <br>操作: {{$issue.AutoFix.Action}}
                    {{end}}
                </p>
                {{if $issue.AutoFix.Code}}
                <p><strong>修复代码:</strong></p>
                <pre><code>{{$issue.AutoFix.Code}}</code></pre>
                {{end}}
                {{end}}
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
