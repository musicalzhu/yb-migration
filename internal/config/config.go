// Package config 提供配置管理功能，负责加载、解析和缓存SQL转换规则。
// 支持从YAML文件加载规则，提供全局配置管理和按类别筛选规则的功能。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// RuleCondition 定义规则匹配的条件，用于在 SQL 文本中定位需要转换或检查的模式。
type RuleCondition struct {
	Pattern string `yaml:"pattern"` // 用于匹配SQL中的模式字符串
}

// RuleAction 定义匹配规则触发后要执行的动作，例如替换函数名或数据类型。
type RuleAction struct {
	Action  string              `yaml:"action"`  // 表示要执行的动作类型
	Target  string              `yaml:"target"`  // 表示动作的目标值
	Mapping []map[string]string `yaml:"mapping"` // 包含从源到目标的映射关系
}

// Rule 表示一条转换或检查规则，包含匹配条件与执行动作。
type Rule struct {
	Name        string        `yaml:"name"`        // 规则的唯一标识符
	Description string        `yaml:"description"` // 描述规则的功能和用途
	Category    string        `yaml:"category"`    // 指定规则所属的类别（function、datatype、syntax、charset）
	When        RuleCondition `yaml:"when"`        // 定义规则匹配的条件
	Then        RuleAction    `yaml:"then"`        // 定义规则匹配后执行的动作
}

// Config 表示加载后的配置文件内容，包含所有规则及元信息。
type Config struct {
	Rules []Rule `yaml:"rules"` // 存储加载的转换规则
	// 新增字段
	LastUpdated string `yaml:"last_updated"` // 最后更新时间
}

// GetRules 返回缓存的规则
func (c *Config) GetRules() []Rule {
	return c.Rules
}

// GetRulesByCategory 按类别筛选规则
func (c *Config) GetRulesByCategory(category string) []Rule {
	var filteredRules []Rule
	for _, rule := range c.Rules {
		if strings.EqualFold(rule.Category, category) {
			filteredRules = append(filteredRules, rule)
		}
	}

	return filteredRules
}

// ResolveFilePath 解析文件路径并验证文件存在
// 参数:
//   - path: 要解析的路径
//   - desc: 路径描述（用于错误信息）
//
// 返回:
//   - string: 绝对路径
//   - error: 错误信息
func ResolveFilePath(path, desc string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%s不能为空", desc)
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("无法解析%s路径'%s': %w", desc, path, err)
	}

	// 检查文件/目录是否存在
	_, err = os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("%s'%s'不存在: %w", desc, absPath, err)
	}

	// 返回绝对路径
	return absPath, nil
}

// GetDefaultConfigPath 返回默认的配置文件路径
// 查找顺序：
// 1. 当前工作目录下的 ./configs/default.yaml（开发时优先）
// 2. 向上查找项目根目录下的 ./configs/default.yaml（测试时兜底）
// 3. 可执行文件所在目录下的 ./configs/default.yaml（部署时兜底）
// 返回:
//   - string: 配置文件路径
//   - error: 错误信息
func GetDefaultConfigPath() (string, error) {
	// 1. 优先使用当前工作目录（开发体验最佳）
	if wd, err := os.Getwd(); err == nil {
		configPath := filepath.Join(wd, "configs", "default.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// 2. 向上查找项目根目录（测试时重要）
	if wd, err := os.Getwd(); err == nil {
		currentDir := wd
		for {
			// 检查当前目录是否有go.mod文件（项目根目录标识）
			goModPath := filepath.Join(currentDir, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				// 找到项目根目录，检查configs目录
				configPath := filepath.Join(currentDir, "configs", "default.yaml")
				if _, err := os.Stat(configPath); err == nil {
					return configPath, nil
				}
				break // 找到项目根目录但没有配置文件
			}

			// 向上一级目录
			parentDir := filepath.Dir(currentDir)
			if parentDir == currentDir {
				break // 已经到达根目录
			}
			currentDir = parentDir
		}
	}

	// 3. 使用可执行文件所在目录（部署时兜底）
	if exe, err := os.Executable(); err == nil {
		configPath := filepath.Join(filepath.Dir(exe), "configs", "default.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	return "", fmt.Errorf("找不到默认配置文件，请显式指定配置文件路径")
}

// GetDefaultReportPath 返回默认的报告输出目录。
// 查找顺序：
// 1. 当前工作目录下的 ./output-report（开发时优先）
// 2. 可执行文件所在目录下的 ./output-report（部署时兜底）
// 该函数永不返回 error，确保程序总能找到一个合理位置。
func GetDefaultReportPath() string {
	// 1. 优先使用当前工作目录（开发体验最佳）
	if wd, err := os.Getwd(); err == nil {
		return filepath.Join(wd, "output-report")
	}

	// 2. 使用可执行文件所在目录（部署时兜底）
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "output-report")
	}

	// 3. 极端情况下返回相对路径（由 filepath.Abs 处理）
	return "output-report"
}

// LoadConfig 加载配置文件
// 参数:
//   - configPath: 配置文件路径，为空时使用默认路径
//
// 返回:
//   - *Config: 配置实例
//   - error: 错误信息
func LoadConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	// 如果没有指定路径，使用默认路径
	if configPath == "" {
		var err error
		configPath, err = GetDefaultConfigPath()
		if err != nil {
			return nil, fmt.Errorf("获取默认配置路径失败: %w", err)
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，返回空配置
		return cfg, nil
	}

	// 读取配置文件
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 YAML
	err = yaml.Unmarshal(file, cfg)
	if err != nil {
		return nil, fmt.Errorf("解析YAML失败: %w", err)
	}

	return cfg, nil
}
