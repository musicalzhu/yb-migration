# 0005. 多格式报告输出

## 状态

接受

## 背景

YB Migration 的用户有不同的需求：
1. 开发者需要 JSON 格式进行程序化处理
2. 项目经理需要 HTML 格式进行可视化查看
3. 文档编写者需要 Markdown 格式集成到文档
4. DBA 需要 SQL 格式直接执行转换后的语句

为了满足不同用户的需求，我们需要支持多种报告格式的输出。

## 决策

实现多格式报告输出系统，支持以下格式：
1. **JSON**: 结构化数据，适合程序化处理
2. **HTML**: 可视化报告，适合人工查看
3. **Markdown**: 文档格式，适合集成到文档系统
4. **SQL**: 转换后的 SQL 语句，适合直接执行

## 后果

### 正面影响

1. **用户友好**: 满足不同用户的需求
2. **工具集成**: 便于与其他工具集成
3. **自动化支持**: JSON 格式支持自动化处理
4. **可视化**: HTML 格式提供直观的问题展示
5. **文档化**: Markdown 格式便于文档编写
6. **实用性**: SQL 格式可以直接使用

### 负面影响

1. **复杂性**: 需要维护多种格式的生成器
2. **性能**: 生成多种格式可能有性能开销
3. **一致性**: 需要确保不同格式内容一致

## 实施细节

### 报告生成器接口

```go
// Reporter 接口定义
type Reporter interface {
    // 生成报告
    Generate(result *AnalysisResult, config *Config) ([]byte, error)
    
    // 获取格式名称
    GetFormat() string
    
    // 获取文件扩展名
    GetExtension() string
    
    // 获取 MIME 类型
    GetMimeType() string
    
    // 获取描述
    GetDescription() string
}

// AnalysisResult 分析结果
type AnalysisResult struct {
    InputPath       string           `json:"input_path"`
    SQLStatements   []SQLStatement   `json:"sql_statements"`
    Issues          []Issue          `json:"issues"`
    Transformations []Transformation `json:"transformations"`
    Summary         AnalysisSummary  `json:"summary"`
    GeneratedAt     time.Time        `json:"generated_at"`
    Version         string           `json:"version"`
}
```

### JSON 报告生成器

```go
type JSONReporter struct{}

func (r *JSONReporter) Generate(result *AnalysisResult, config *Config) ([]byte, error) {
    result.GeneratedAt = time.Now()
    result.Version = "2.0.0"
    
    return json.MarshalIndent(result, "", "  ")
}

func (r *JSONReporter) GetFormat() string {
    return "JSON"
}

func (r *JSONReporter) GetExtension() string {
    return ".json"
}

func (r *JSONReporter) GetMimeType() string {
    return "application/json"
}

func (r *JSONReporter) GetDescription() string {
    return "结构化 JSON 格式，适合程序化处理"
}
```

### HTML 报告生成器

```go
type HTMLReporter struct {
    template *template.Template
}

func NewHTMLReporter() (*HTMLReporter, error) {
    tmpl, err := template.New("report").Parse(htmlTemplate)
    if err != nil {
        return nil, err
    }
    
    return &HTMLReporter{template: tmpl}, nil
}

func (r *HTMLReporter) Generate(result *AnalysisResult, config *Config) ([]byte, error) {
    result.GeneratedAt = time.Now()
    result.Version = "2.0.0"
    
    var buf bytes.Buffer
    err := r.template.Execute(&buf, map[string]interface{}{
        "Result": result,
        "Config": config,
    })
    
    return buf.Bytes(), err
}

func (r *HTMLReporter) GetFormat() string {
    return "HTML"
}

func (r *HTMLReporter) GetExtension() string {
    return ".html"
}

func (r *HTMLReporter) GetMimeType() string {
    return "text/html"
}

func (r *HTMLReporter) GetDescription() string {
    return "可视化 HTML 报告，适合人工查看"
}

// HTML 模板
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>YB Migration 分析报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .issue { margin: 10px 0; padding: 10px; border-left: 4px solid #ccc; }
        .error { border-left-color: #d32f2f; }
        .warning { border-left-color: #f57c00; }
        .info { border-left-color: #1976d2; }
        .summary { background: #e3f2fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YB Migration 分析报告</h1>
        <p><strong>输入文件:</strong> {{.Result.InputPath}}</p>
        <p><strong>生成时间:</strong> {{.Result.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
        <p><strong>版本:</strong> {{.Result.Version}}</p>
    </div>
    
    <div class="summary">
        <h2>分析摘要</h2>
        <table>
            <tr><th>总语句数</th><td>{{.Result.Summary.TotalStatements}}</td></tr>
            <tr><th>总问题数</th><td>{{.Result.Summary.TotalIssues}}</td></tr>
            <tr><th>错误</th><td>{{.Result.Summary.IssuesBySeverity.error}}</td></tr>
            <tr><th>警告</th><td>{{.Result.Summary.IssuesBySeverity.warning}}</td></tr>
            <tr><th>信息</th><td>{{.Result.Summary.IssuesBySeverity.info}}</td></tr>
        </table>
    </div>
    
    <h2>发现的问题</h2>
    {{range .Result.Issues}}
    <div class="issue {{.Severity}}">
        <h3>{{.Checker}} - {{.Severity}}</h3>
        <p><strong>消息:</strong> {{.Message}}</p>
        <p><strong>位置:</strong> 第 {{.LineNumber}} 行，第 {{.Column}} 列</p>
        {{if .Suggestion}}<p><strong>建议:</strong> {{.Suggestion}}</p>{{end}}
    </div>
    {{end}}
    
    <h2>转换建议</h2>
    {{range .Result.Transformations}}
    <div class="transformation">
        <h3>转换建议</h3>
        <p><strong>原始:</strong> <code>{{.Original}}</code></p>
        <p><strong>转换后:</strong> <code>{{.Transformed}}</code></p>
        <p><strong>原因:</strong> {{.Reason}}</p>
    </div>
    {{end}}
</body>
</html>
`
```

### Markdown 报告生成器

```go
type MarkdownReporter struct{}

func (r *MarkdownReporter) Generate(result *AnalysisResult, config *Config) ([]byte, error) {
    result.GeneratedAt = time.Now()
    result.Version = "2.0.0"
    
    var buf bytes.Buffer
    
    // 标题
    buf.WriteString("# YB Migration 分析报告\n\n")
    
    // 基本信息
    buf.WriteString("## 基本信息\n\n")
    buf.WriteString(fmt.Sprintf("- **输入文件**: %s\n", result.InputPath))
    buf.WriteString(fmt.Sprintf("- **生成时间**: %s\n", result.GeneratedAt.Format("2006-01-02 15:04:05")))
    buf.WriteString(fmt.Sprintf("- **版本**: %s\n\n", result.Version))
    
    // 摘要
    buf.WriteString("## 分析摘要\n\n")
    buf.WriteString("| 指标 | 数量 |\n")
    buf.WriteString("|------|------|\n")
    buf.WriteString(fmt.Sprintf("| 总语句数 | %d |\n", result.Summary.TotalStatements))
    buf.WriteString(fmt.Sprintf("| 总问题数 | %d |\n", result.Summary.TotalIssues))
    buf.WriteString(fmt.Sprintf("| 错误 | %d |\n", result.Summary.IssuesBySeverity.error))
    buf.WriteString(fmt.Sprintf("| 警告 | %d |\n", result.Summary.IssuesBySeverity.warning))
    buf.WriteString(fmt.Sprintf("| 信息 | %d |\n\n", result.Summary.IssuesBySeverity.info))
    
    // 问题列表
    if len(result.Issues) > 0 {
        buf.WriteString("## 发现的问题\n\n")
        for i, issue := range result.Issues {
            buf.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, issue.Checker))
            buf.WriteString(fmt.Sprintf("- **严重程度**: %s\n", issue.Severity))
            buf.WriteString(fmt.Sprintf("- **消息**: %s\n", issue.Message))
            buf.WriteString(fmt.Sprintf("- **位置**: 第 %d 行，第 %d 列\n", issue.LineNumber, issue.Column))
            if issue.Suggestion != "" {
                buf.WriteString(fmt.Sprintf("- **建议**: %s\n", issue.Suggestion))
            }
            buf.WriteString("\n")
        }
    }
    
    // 转换建议
    if len(result.Transformations) > 0 {
        buf.WriteString("## 转换建议\n\n")
        for i, trans := range result.Transformations {
            buf.WriteString(fmt.Sprintf("### %d. 转换建议\n\n", i+1))
            buf.WriteString(fmt.Sprintf("**原始**:\n```sql\n%s\n```\n\n", trans.Original))
            buf.WriteString(fmt.Sprintf("**转换后**:\n```sql\n%s\n```\n\n", trans.Transformed))
            buf.WriteString(fmt.Sprintf("**原因**: %s\n\n", trans.Reason))
        }
    }
    
    return buf.Bytes(), nil
}

func (r *MarkdownReporter) GetFormat() string {
    return "Markdown"
}

func (r *MarkdownReporter) GetExtension() string {
    return ".md"
}

func (r *MarkdownReporter) GetMimeType() string {
    return "text/markdown"
}

func (r *MarkdownReporter) GetDescription() string {
    return "Markdown 文档格式，适合集成到文档系统"
}
```

### SQL 报告生成器

```go
type SQLReporter struct{}

func (r *SQLReporter) Generate(result *AnalysisResult, config *Config) ([]byte, error) {
    var buf bytes.Buffer
    
    // 文件头注释
    buf.WriteString("-- YB Migration 转换后的 SQL\n")
    buf.WriteString(fmt.Sprintf("-- 原文件: %s\n", result.InputPath))
    buf.WriteString(fmt.Sprintf("-- 生成时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
    buf.WriteString(fmt.Sprintf("-- 版本: %s\n\n", "2.0.0"))
    
    // 输出转换后的 SQL
    for _, trans := range result.Transformations {
        buf.WriteString("-- 转换原因: ")
        buf.WriteString(trans.Reason)
        buf.WriteString("\n")
        buf.WriteString("-- 原始: ")
        buf.WriteString(trans.Original)
        buf.WriteString("\n")
        buf.WriteString(trans.Transformed)
        buf.WriteString("\n\n")
    }
    
    return buf.Bytes(), nil
}

func (r *SQLReporter) GetFormat() string {
    return "SQL"
}

func (r *SQLReporter) GetExtension() string {
    return ".sql"
}

func (r *SQLReporter) GetMimeType() string {
    return "application/sql"
}

func (r *SQLReporter) GetDescription() string {
    return "转换后的 SQL 语句，适合直接执行"
}
```

### 报告管理器

```go
type ReportManager struct {
    reporters map[string]Reporter
}

func NewReportManager() *ReportManager {
    rm := &ReportManager{
        reporters: make(map[string]Reporter),
    }
    
    // 注册内置报告生成器
    rm.RegisterReporter("json", &JSONReporter{})
    
    htmlReporter, _ := NewHTMLReporter()
    rm.RegisterReporter("html", htmlReporter)
    
    rm.RegisterReporter("markdown", &MarkdownReporter{})
    rm.RegisterReporter("sql", &SQLReporter{})
    
    return rm
}

func (rm *ReportManager) RegisterReporter(format string, reporter Reporter) {
    rm.reporters[format] = reporter
}

func (rm *ReportManager) GenerateReports(result *AnalysisResult, config *Config, outputDir string) error {
    for _, format := range config.Reports.Formats {
        reporter, exists := rm.reporters[format]
        if !exists {
            return fmt.Errorf("不支持的报告格式: %s", format)
        }
        
        data, err := reporter.Generate(result, config)
        if err != nil {
            return fmt.Errorf("生成 %s 报告失败: %w", format, err)
        }
        
        filename := "summary" + reporter.GetExtension()
        filepath := path.Join(outputDir, filename)
        
        err = ioutil.WriteFile(filepath, data, 0644)
        if err != nil {
            return fmt.Errorf("写入报告文件失败: %w", err)
        }
    }
    
    return nil
}
```

## 替代方案

### 单一格式输出
- **优点**: 实现简单
- **缺点**: 不能满足不同用户需求

### 插件化报告生成器
- **优点**: 极高扩展性
- **缺点**: 过度复杂，不适合当前需求

### 配置驱动模板
- **优点**: 灵活配置
- **缺点**: 实现复杂，性能较差

## 相关决策

- [0002. 模块化架构设计](0002-modular-architecture.md)
- [0004. 插件化检查器架构](0004-plugin-checkers.md)

## 参考资料

- [Go 模板引擎](https://golang.org/pkg/text/template/)
- [Markdown 规范](https://commonmark.org/)
- [HTML5 规范](https://html.spec.whatwg.org/)

---

*创建日期: 2026-02-03*  
*最后更新: 2026-02-03*  
*负责人: YB Migration Team*
