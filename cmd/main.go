// Package main 是 CLI 程序入口，负责解析命令行参数、加载配置，并
// 调用分析与报告生成功能。可通过 `run` 函数进行单元测试。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/ybMigration/internal/analyzer"
	"github.com/example/ybMigration/internal/config"
	report "github.com/example/ybMigration/internal/report"
	sqlparser "github.com/example/ybMigration/internal/sql-parser"
)

// 定义退出状态码
const (
	exitSuccess     = 0 // 成功
	exitInvalidArgs = 1 // 参数错误
	exitConfigError = 2 // 配置错误
	exitAnalysisErr = 3 // 分析错误
)

// run 是主要的执行函数，可以被测试
// 返回错误而不是直接退出程序
func run(absConfigPath, absPath, absReportPath string) error {

	// 验证配置文件存在
	if _, err := config.ResolveFilePath(absConfigPath, "配置文件"); err != nil {
		return fmt.Errorf("配置文件验证失败: %w", err)
	}

	// 创建分析器工厂
	af, err := analyzer.NewAnalyzerFactory(absConfigPath)
	if err != nil {
		return fmt.Errorf("创建分析器工厂失败: %w", err)
	}

	// 创建检查器
	checkers, err := af.CreateCheckers()
	if err != nil {
		return fmt.Errorf("创建检查器失败: %w", err)
	}

	// 创建 SQL 解析器
	sqlParser := sqlparser.NewSQLParser()
	if sqlParser == nil {
		return fmt.Errorf("创建SQL解析器失败")
	}

	// 使用分析器分析输入
	result, err := analyzer.AnalyzeInput(absPath, sqlParser, checkers)
	if err != nil {
		return fmt.Errorf("分析输入失败: %w", err)
	}

	// 生成所有支持格式的报告
	if err := report.GenerateReports(absReportPath, result); err != nil {
		return fmt.Errorf("生成报告失败: %w", err)
	}

	fmt.Println("分析完成！报告已生成")

	return nil
}

func main() {
	absConfigPath, absPath, absReportPath := parseFlags()

	if err := run(absConfigPath, absPath, absReportPath); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(exitAnalysisErr)
	}
}

// parseFlags 解析命令行参数，返回绝对路径的配置文件、分析路径和报告目录。
func parseFlags() (absConfigPath, absPath, absReportPath string) {
	var configPath, path, reportPath string

	// 定义命令行参数
	flag.StringVar(&configPath, "config", "", "配置文件路径（YAML）。若未指定，则自动查找默认位置。")
	flag.StringVar(&path, "path", "", "待分析的SQL文件、日志文件或目录路径（可选，若未提供则从 args[0] 参数读取）。")
	flag.StringVar(&reportPath, "reportPath", "", "分析报告输出目录。若未指定，默认为 ./reports。")

	// 自定义帮助信息
	flag.Usage = func() {
		exe := os.Args[0]
		fmt.Fprintf(os.Stderr, "用法:\n")
		fmt.Fprintf(os.Stderr, "  %s [选项] <路径>\n", exe)
		fmt.Fprintf(os.Stderr, "  %s --config <配置文件> --path <路径> --reportPath <报告目录>\n\n", exe)
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// --- 解析待分析路径（必须提供）---
	var finalPath string
	if path != "" {
		finalPath = path
	} else {
		args := flag.Args()
		switch len(args) {
		case 1:
			finalPath = args[0]
		case 0:
			fmt.Fprintf(os.Stderr, "错误: 未指定待分析路径（可通过 --path 或直接传参）\n")
			flag.Usage()
			os.Exit(1)
		default:
			fmt.Fprintf(os.Stderr, "错误: 仅支持一个 positional 参数（SQL文件、日志文件或目录路径）\n")
			flag.Usage()
			os.Exit(1)
		}
	}

	var err error
	// 解析待分析路径（必须存在）
	absPath, err = config.ResolveFilePath(finalPath, "待分析路径")
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	// --- 解析配置文件路径（可选）---
	if configPath != "" {
		absConfigPath, err = config.ResolveFilePath(configPath, "配置文件")
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 使用默认配置查找逻辑
		absConfigPath, err = config.GetDefaultConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 未指定配置文件，且默认查找失败：%v\n", err)
			os.Exit(1)
		}
	}

	// --- 解析报告输出目录（可选）---
	if reportPath != "" {
		// 用户显式指定：转为绝对路径
		absReportPath, err = filepath.Abs(reportPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无法解析报告目录路径 '%s': %v\n", reportPath, err)
			os.Exit(1)
		}
	} else {
		// 使用默认报告目录逻辑
		defaultReportPath := config.GetDefaultReportPath()
		absReportPath, err = filepath.Abs(defaultReportPath)
		if err != nil {
			// 理论上不会发生，但防御性处理
			fmt.Fprintf(os.Stderr, "错误: 无法确定报告目录路径: %v\n", err)
			os.Exit(1)
		}
	}

	// 确保报告目录存在（自动创建）
	if err := os.MkdirAll(absReportPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无法创建报告目录 %s: %v\n", absReportPath, err)
		os.Exit(1)
	}

	// 打印确认信息
	fmt.Printf("✅ 使用配置文件: %s\n", absConfigPath)
	fmt.Printf("✅ 待分析路径: %s\n", absPath)
	fmt.Printf("✅ 报告输出目录: %s\n", absReportPath)
	return
}
