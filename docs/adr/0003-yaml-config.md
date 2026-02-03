# 0003. 使用 YAML 配置文件

## 状态

接受

## 背景

YB Migration 需要支持灵活的配置管理，包括：
1. 检查规则的启用/禁用
2. 检查参数的调整
3. 报告格式的选择
4. 输出路径的配置

我们需要一个人类可读、易于编辑、支持复杂数据结构的配置格式。

## 决策

使用 YAML 格式作为配置文件格式，支持以下特性：
1. 层次化配置结构
2. 注释支持
3. 多环境配置
4. 配置继承和覆盖

## 后果

### 正面影响

1. **人类可读**: YAML 格式清晰易读，支持注释
2. **层次结构**: 支持复杂的嵌套配置
3. **广泛支持**: Go 有优秀的 YAML 库支持
4. **易于编辑**: 可以用任何文本编辑器编辑
5. **版本控制友好**: 文本格式，易于版本控制
6. **扩展性**: 易于添加新的配置项

### 负面影响

1. **格式敏感**: 缩进错误会导致解析失败
2. **性能**: 解析 YAML 比 JSON 稍慢
3. **复杂性**: 复杂配置可能难以理解

## 实施细节

### 配置文件结构

```yaml
# YB Migration 配置文件
rules:
  # 函数兼容性检查
  - name: "function_incompatibility"
    category: "function"
    description: "检查不兼容的函数调用"
    enabled: true
    severity: "error"
    parameters:
      target_database: "yugabytedb"
      strict_mode: true

  # 数据类型检查
  - name: "datatype_incompatibility"
    category: "datatype"
    description: "检查不兼容的数据类型"
    enabled: true
    severity: "warning"
    parameters:
      allow_auto_conversion: false

  # 语法检查
  - name: "syntax_incompatibility"
    category: "syntax"
    description: "检查语法兼容性"
    enabled: true
    severity: "error"
    parameters:
      mysql_version: "8.0"
      yugabytedb_version: "2.0"

  # 字符集检查
  - name: "charset_incompatibility"
    category: "charset"
    description: "检查字符集兼容性"
    enabled: true
    severity: "info"
    parameters:
      default_charset: "utf8mb4"
      target_charset: "utf8"

# 报告配置
reports:
  formats:
    - "json"
    - "html"
    - "markdown"
    - "sql"
  
  output:
    directory: "./output-report"
    include_transformed_sql: true
    include_summary: true

# 全局设置
global:
  log_level: "info"
  parallel_processing: true
  max_workers: 4
  timeout: "30s"
```

### 配置文件查找顺序

1. 命令行指定的配置文件
2. `./config.yaml`
3. `~/.yb-migration/config.yaml`
4. `/etc/yb-migration/config.yaml`

### 配置加载实现

```go
type Config struct {
    Rules  []Rule  `yaml:"rules"`
    Reports Reports `yaml:"reports"`
    Global  Global  `yaml:"global"`
}

type Rule struct {
    Name        string                 `yaml:"name"`
    Category    string                 `yaml:"category"`
    Description string                 `yaml:"description"`
    Enabled     bool                   `yaml:"enabled"`
    Severity    string                 `yaml:"severity"`
    Parameters  map[string]interface{} `yaml:"parameters"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## 替代方案

### JSON 配置
- **优点**: 解析速度快，格式严格
- **缺点**: 不支持注释，可读性较差

### TOML 配置
- **优点**: 语法简洁，支持注释
- **缺点**: 生态系统相对较小

### 环境变量
- **优点**: 容器友好
- **缺点**: 不适合复杂配置，可读性差

### 数据库配置
- **优点**: 动态配置，多实例共享
- **缺点**: 过度复杂，不适合 CLI 工具

## 相关决策

- [0002. 模块化架构设计](0002-modular-architecture.md)
- [0004. 插件化检查器架构](0004-plugin-checkers.md)

## 参考资料

- [YAML 官方文档](https://yaml.org/)
- [Go YAML 库](https://github.com/go-yaml/yaml)
- [配置文件最佳实践](https://12factor.net/config)

---

*创建日期: 2026-02-03*  
*最后更新: 2026-02-03*  
*负责人: YB Migration Team*
