package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/example/ybMigration/internal/constants"
	"github.com/example/ybMigration/internal/model"
)

// JSONGenerator 生成 JSON 格式的报告
type JSONGenerator struct{}

// Write 将报告写入 JSON 文件
func (g *JSONGenerator) Write(path string, report model.Report) error {
	// 使用缩进格式化 JSON 输出
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, constants.FilePermission); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
