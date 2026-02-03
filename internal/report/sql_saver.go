// Package report 提供报告生成功能
package report

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/ybMigration/internal/constants"
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
	if err := os.MkdirAll(dir, constants.DirPermission); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建文件（提前声明 err 用于聚合）
	file, err := os.Create(outputPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer func() {
		// 仅当主流程无错误时，才用 Close() 错误覆盖返回值
		// （因已 Sync()，Close() 错误通常不关键，但保留诊断价值）
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("关闭文件时发生次要错误（数据已安全落盘）: %w", closeErr)
		}
	}()

	// 写入内容
	_, err = file.WriteString(result.TransformedSQL)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 确保数据写入磁盘（关键步骤）
	if err = file.Sync(); err != nil {
		return fmt.Errorf("同步文件失败: %w", err)
	}

	return nil // 此时 err == nil，Close() 错误会在此处被捕获
}
