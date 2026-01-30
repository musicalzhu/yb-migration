// Package testutils 提供测试工具函数
package testutils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/example/ybMigration/internal/config"
)

// GetTestDataPath 返回测试数据文件或目录的完整路径
// 参数:
//   - path: 相对于 testdata 目录的文件或目录路径，如果为空则返回 testdata 目录的路径
//
// 返回:
//   - 完整的文件系统路径
//   - 错误信息（如果路径无效）
func GetTestDataPath(path string) (string, error) {
	// 获取当前文件的路径
	_, currentFile, _, ok := runtime.Caller(1) // 使用调用栈上一级
	if !ok {
		return "", os.ErrNotExist
	}

	// 计算项目根目录 - 向上查找直到找到 go.mod
	baseDir := filepath.Dir(filepath.Dir(currentFile))
	for {
		goModPath := filepath.Join(baseDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break
		}

		parent := filepath.Dir(baseDir)
		if parent == baseDir {
			break // 已经到达根目录
		}
		baseDir = parent
	}

	testDataDir := filepath.Join(baseDir, "testdata")

	// 如果路径为空，返回测试数据目录
	if path == "" {
		return testDataDir, nil
	}

	// 构建完整路径
	fullPath := filepath.Join(testDataDir, path)

	return fullPath, nil
}

// MustGetTestDataPath 是 GetTestDataPath 的便捷包装，如果出错会 panic
// 适用于测试初始化等场景
func MustGetTestDataPath(path string) string {
	p, err := GetTestDataPath(path)
	if err != nil {
		panic(err)
	}
	return p
}

// ============================================================================
// 测试配置辅助
// ============================================================================

// TestConfig 测试配置单例
var TestConfig *config.Config

// SetupTestConfig 设置测试配置（在所有测试前调用一次）
func SetupTestConfig(t *testing.T) *config.Config {
	if TestConfig != nil {
		return TestConfig
	}

	// 直接使用config包的LoadConfig方法加载默认配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		if t != nil {
			t.Fatalf("创建测试配置失败: %v", err)
		}
		panic(err)
	}

	TestConfig = cfg
	if t != nil {
		t.Logf("测试配置初始化完成，加载了 %d 个规则", len(cfg.Rules))
	}

	return cfg
}

// GetTestConfig 获取测试配置
func GetTestConfig(t *testing.T) *config.Config {
	if TestConfig == nil {
		SetupTestConfig(t)
	}
	return TestConfig
}

// ResetTestConfig 重置测试配置（用于测试隔离）
func ResetTestConfig() {
	TestConfig = nil
}
