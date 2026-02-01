package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 结构体测试
// ============================================================================

func TestRuleCondition(t *testing.T) {
	// 测试 RuleCondition 结构体
	t.Run("valid_condition", func(t *testing.T) {
		condition := RuleCondition{
			Pattern: "GROUP_CONCAT",
		}
		assert.Equal(t, "GROUP_CONCAT", condition.Pattern)
	})
}

func TestRuleAction(t *testing.T) {
	// 测试 RuleAction 结构体
	t.Run("valid_action", func(t *testing.T) {
		action := RuleAction{
			Action: "replace",
			Target: "STRING_AGG",
			Mapping: []map[string]string{
				{"GROUP_CONCAT": "STRING_AGG"},
			},
		}
		assert.Equal(t, "replace", action.Action)
		assert.Equal(t, "STRING_AGG", action.Target)
		assert.Len(t, action.Mapping, 1)
	})
}

func TestRule(t *testing.T) {
	// 测试 Rule 结构体
	t.Run("valid_rule", func(t *testing.T) {
		rule := Rule{
			Name:        "GROUP_CONCAT_to_STRING_AGG",
			Description: "将MySQL的GROUP_CONCAT函数转换为YB兼容的STRING_AGG函数",
			Category:    "function",
			When:        RuleCondition{Pattern: "GROUP_CONCAT"},
			Then: RuleAction{
				Action: "replace",
				Target: "STRING_AGG",
			},
		}
		assert.Equal(t, "GROUP_CONCAT_to_STRING_AGG", rule.Name)
		assert.Equal(t, "function", rule.Category)
		assert.Equal(t, "GROUP_CONCAT", rule.When.Pattern)
		assert.Equal(t, "replace", rule.Then.Action)
	})
}

func TestConfig(t *testing.T) {
	// 测试 Config 结构体
	t.Run("empty_config", func(t *testing.T) {
		cfg := &Config{}
		assert.Empty(t, cfg.Rules)
		assert.Empty(t, cfg.LastUpdated)
	})

	t.Run("config_with_rules", func(t *testing.T) {
		rules := []Rule{
			{Name: "rule1", Category: "function"},
			{Name: "rule2", Category: "datatype"},
		}
		cfg := &Config{
			Rules:       rules,
			LastUpdated: "2024-01-01",
		}
		assert.Len(t, cfg.Rules, 2)
		assert.Equal(t, "2024-01-01", cfg.LastUpdated)
	})
}

// ============================================================================
// Config 方法测试
// ============================================================================

func TestConfig_GetRules(t *testing.T) {
	t.Run("empty_rules", func(t *testing.T) {
		cfg := &Config{Rules: []Rule{}}
		rules := cfg.GetRules()
		assert.Empty(t, rules)
	})

	t.Run("multiple_rules", func(t *testing.T) {
		rules := []Rule{
			{Name: "rule1", Category: "function"},
			{Name: "rule2", Category: "datatype"},
		}
		cfg := &Config{Rules: rules}
		result := cfg.GetRules()
		assert.Equal(t, rules, result)
	})
}

func TestConfig_GetRulesByCategory(t *testing.T) {
	rules := []Rule{
		{Name: "rule1", Category: "function"},
		{Name: "rule2", Category: "datatype"},
		{Name: "rule3", Category: "function"},
		{Name: "rule4", Category: "syntax"},
	}
	cfg := &Config{Rules: rules}

	t.Run("filter_function", func(t *testing.T) {
		functionRules := cfg.GetRulesByCategory("function")
		assert.Len(t, functionRules, 2)
		assert.Equal(t, "rule1", functionRules[0].Name)
		assert.Equal(t, "rule3", functionRules[1].Name)
	})

	t.Run("filter_datatype", func(t *testing.T) {
		datatypeRules := cfg.GetRulesByCategory("datatype")
		assert.Len(t, datatypeRules, 1)
		assert.Equal(t, "rule2", datatypeRules[0].Name)
	})

	t.Run("filter_case_insensitive", func(t *testing.T) {
		functionRules := cfg.GetRulesByCategory("FUNCTION")
		assert.Len(t, functionRules, 2)
	})

	t.Run("filter_nonexistent", func(t *testing.T) {
		nonexistentRules := cfg.GetRulesByCategory("nonexistent")
		assert.Empty(t, nonexistentRules)
	})
}

// ============================================================================
// 工具函数测试
// ============================================================================

func TestResolveFilePath(t *testing.T) {
	t.Run("empty_path", func(t *testing.T) {
		_, err := ResolveFilePath("", "测试文件")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不能为空")
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		_, err := ResolveFilePath("/nonexistent/file.yaml", "测试文件")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不存在")
	})

	t.Run("valid_file", func(t *testing.T) {
		// 创建临时文件
		tmpFile := filepath.Join(t.TempDir(), "test.yaml")
		err := os.WriteFile(tmpFile, []byte("test: content"), 0644)
		require.NoError(t, err)

		resolvedPath, err := ResolveFilePath(tmpFile, "测试文件")
		assert.NoError(t, err)
		assert.Equal(t, tmpFile, resolvedPath)
	})

	t.Run("relative_path", func(t *testing.T) {
		// 创建临时文件
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.yaml")
		err := os.WriteFile(tmpFile, []byte("test: content"), 0644)
		require.NoError(t, err)

		// 切换到临时目录
		originalDir, _ := os.Getwd()
		defer func() {
			if err := os.Chdir(originalDir); err != nil {
				t.Logf("恢复工作目录失败: %v", err)
			}
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		resolvedPath, err := ResolveFilePath("test.yaml", "测试文件")
		assert.NoError(t, err)
		assert.True(t, filepath.IsAbs(resolvedPath))
	})
}

func TestGetDefaultConfigPath(t *testing.T) {
	t.Run("no_config_found", func(t *testing.T) {
		// 保存原始工作目录
		originalDir, _ := os.Getwd()
		defer func() {
			if err := os.Chdir(originalDir); err != nil {
				t.Logf("恢复工作目录失败: %v", err)
			}
		}()

		// 创建空目录
		tmpDir := t.TempDir()
		err := os.Chdir(tmpDir)
		require.NoError(t, err)

		// 测试获取默认配置路径
		_, err = GetDefaultConfigPath()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "找不到默认配置文件")
	})
}

func TestGetDefaultReportPath(t *testing.T) {
	t.Run("basic_functionality", func(t *testing.T) {
		// 测试函数不会 panic
		reportPath := GetDefaultReportPath()
		assert.NotEmpty(t, reportPath)
		assert.Contains(t, reportPath, "output-report")
	})

	t.Run("use_working_directory", func(t *testing.T) {
		originalWD, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalWD)
		}()

		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		reportPath := GetDefaultReportPath()
		assert.Equal(t, filepath.Join(tmpDir, "output-report"), reportPath)
	})
}

// ============================================================================
// LoadConfig 测试
// ============================================================================

func TestLoadConfig(t *testing.T) {
	t.Run("load_default_config", func(t *testing.T) {
		// 这个测试使用项目的默认配置文件
		cfg, err := LoadConfig("")
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.NotEmpty(t, cfg.Rules)
	})

	t.Run("load_nonexistent_file", func(t *testing.T) {
		cfg, err := LoadConfig("/nonexistent/config.yaml")
		assert.NoError(t, err) // 应该返回空配置而不是错误
		assert.NotNil(t, cfg)
		assert.Empty(t, cfg.Rules)
	})

	t.Run("load_custom_config", func(t *testing.T) {
		// 创建临时配置文件
		tmpFile := filepath.Join(t.TempDir(), "test.yaml")
		configContent := `
rules:
  - name: "test_rule"
    description: "测试规则"
    category: "function"
    when:
      pattern: "TEST_FUNC"
    then:
      action: "replace"
      target: "REPLACED_FUNC"
last_updated: "2024-01-01"
`
		err := os.WriteFile(tmpFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 加载配置
		cfg, err := LoadConfig(tmpFile)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Len(t, cfg.Rules, 1)
		assert.Equal(t, "test_rule", cfg.Rules[0].Name)
		assert.Equal(t, "function", cfg.Rules[0].Category)
		assert.Equal(t, "TEST_FUNC", cfg.Rules[0].When.Pattern)
		assert.Equal(t, "replace", cfg.Rules[0].Then.Action)
		assert.Equal(t, "REPLACED_FUNC", cfg.Rules[0].Then.Target)
		assert.Equal(t, "2024-01-01", cfg.LastUpdated)
	})

	t.Run("load_invalid_yaml", func(t *testing.T) {
		// 创建无效的YAML文件
		tmpFile := filepath.Join(t.TempDir(), "invalid.yaml")
		invalidContent := `
rules:
  - name: "test_rule"
    description: "测试规则"
    invalid_yaml: [
`
		err := os.WriteFile(tmpFile, []byte(invalidContent), 0644)
		require.NoError(t, err)

		// 加载配置应该失败
		cfg, err := LoadConfig(tmpFile)
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "解析YAML失败")
	})

	t.Run("load_empty_file", func(t *testing.T) {
		// 创建空文件
		tmpFile := filepath.Join(t.TempDir(), "empty.yaml")
		err := os.WriteFile(tmpFile, []byte(""), 0644)
		require.NoError(t, err)

		// 加载配置
		cfg, err := LoadConfig(tmpFile)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Empty(t, cfg.Rules)
	})

	t.Run("load_file_without_rules", func(t *testing.T) {
		// 创建没有rules字段的文件
		tmpFile := filepath.Join(t.TempDir(), "no_rules.yaml")
		content := `
last_updated: "2024-01-01"
other_field: "value"
`
		err := os.WriteFile(tmpFile, []byte(content), 0644)
		require.NoError(t, err)

		// 加载配置
		cfg, err := LoadConfig(tmpFile)
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Empty(t, cfg.Rules)
		assert.Equal(t, "2024-01-01", cfg.LastUpdated)
	})
}

// ============================================================================
// 集成测试
// ============================================================================

func TestConfigIntegration(t *testing.T) {
	t.Run("full_workflow", func(t *testing.T) {
		// 创建完整的配置文件
		tmpFile := filepath.Join(t.TempDir(), "integration.yaml")
		configContent := `
rules:
  - name: "GROUP_CONCAT_to_STRING_AGG"
    description: "将MySQL的GROUP_CONCAT函数转换为YB兼容的STRING_AGG函数"
    category: "function"
    when:
      pattern: "GROUP_CONCAT"
    then:
      action: "replace"
      target: "STRING_AGG"
  
  - name: "TINYINT_to_SMALLINT"
    description: "将TINYINT数据类型转换为SMALLINT"
    category: "datatype"
    when:
      pattern: "TINYINT"
    then:
      action: "replace"
      target: "SMALLINT"
  
  - name: "AUTO_INCREMENT_to_SERIAL"
    description: "将MySQL的AUTO_INCREMENT转换为PostgreSQL兼容的SERIAL"
    category: "syntax"
    when:
      pattern: "AUTO_INCREMENT"
    then:
      action: "replace"
      target: "SERIAL"
last_updated: "2024-01-01T00:00:00Z"
`
		err := os.WriteFile(tmpFile, []byte(configContent), 0644)
		require.NoError(t, err)

		// 加载配置
		cfg, err := LoadConfig(tmpFile)
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// 验证规则数量
		assert.Len(t, cfg.Rules, 3)
		assert.Equal(t, "2024-01-01T00:00:00Z", cfg.LastUpdated)

		// 验证按类别筛选
		functionRules := cfg.GetRulesByCategory("function")
		assert.Len(t, functionRules, 1)
		assert.Equal(t, "GROUP_CONCAT_to_STRING_AGG", functionRules[0].Name)

		datatypeRules := cfg.GetRulesByCategory("datatype")
		assert.Len(t, datatypeRules, 1)
		assert.Equal(t, "TINYINT_to_SMALLINT", datatypeRules[0].Name)

		syntaxRules := cfg.GetRulesByCategory("syntax")
		assert.Len(t, syntaxRules, 1)
		assert.Equal(t, "AUTO_INCREMENT_to_SERIAL", syntaxRules[0].Name)

		// 验证规则内容
		rule := cfg.Rules[0]
		assert.Equal(t, "GROUP_CONCAT_to_STRING_AGG", rule.Name)
		assert.Equal(t, "将MySQL的GROUP_CONCAT函数转换为YB兼容的STRING_AGG函数", rule.Description)
		assert.Equal(t, "function", rule.Category)
		assert.Equal(t, "GROUP_CONCAT", rule.When.Pattern)
		assert.Equal(t, "replace", rule.Then.Action)
		assert.Equal(t, "STRING_AGG", rule.Then.Target)
	})
}
