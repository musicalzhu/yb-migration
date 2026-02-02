// Package report 提供报告生成功能
package report

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/ybMigration/internal/model"
)

// GenerateTransformedSQLPath 生成转换后 SQL 文件的路径
// sourcePath: 源文件路径
// outputDir: 输出目录
// 返回值: 转换后的SQL文件路径
func GenerateTransformedSQLPath(sourcePath, outputDir string) string {
	// 获取文件名（不含扩展名）
	baseName := filepath.Base(sourcePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]
	
	// 构造输出文件名
	outputFileName := fmt.Sprintf("%s_transformed.sql", nameWithoutExt)
	return filepath.Join(outputDir, outputFileName)
}

// SaveTransformedSQL 保存转换后的SQL到文件
// result: 分析结果
// outputPath: 输出文件路径
// 返回值: 错误信息
func SaveTransformedSQL(result model.AnalysisResult, outputPath string) error {
	if result.TransformedSQL == "" {
		return fmt.Errorf("没有转换后的SQL需要保存")
	}

	// 确保输出目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建文件
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入内容
	_, err = file.WriteString(result.TransformedSQL)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 确保数据写入磁盘
	if err := file.Sync(); err != nil {
		return fmt.Errorf("同步文件失败: %w", err)
	}

	return nil
}

// SaveMultipleTransformedSQL 批量保存多个转换后的SQL文件
// results: 分析结果列表
// sourcePaths: 对应的源文件路径列表
// outputDir: 输出目录
// 返回值: 错误信息
func SaveMultipleTransformedSQL(results []model.AnalysisResult, sourcePaths []string, outputDir string) error {
	if len(results) != len(sourcePaths) {
		return fmt.Errorf("结果数量与路径数量不匹配")
	}

	for i, result := range results {
		if result.TransformedSQL == "" {
			continue // 跳过没有转换内容的文件
		}

		outputPath := GenerateTransformedSQLPath(sourcePaths[i], outputDir)
		if err := SaveTransformedSQL(result, outputPath); err != nil {
			return fmt.Errorf("保存文件 %s 失败: %w", outputPath, err)
		}
	}

	return nil
}
