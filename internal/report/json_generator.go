package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/example/ybMigration/internal/model"
)

// JSONGenerator 生成 JSON 格式的报告
type JSONGenerator struct{}

// Write 将分析结果写入 JSON 文件
func (g *JSONGenerator) Write(path string, result model.AnalysisResult) error {
	// 使用缩进格式化 JSON 输出
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON 编码失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}
