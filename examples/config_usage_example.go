// Package main 提供示例程序，演示如何加载和使用配置。
// 该示例仅用于开发与文档目的，不用于生产环境。
package main

import (
	"fmt"
	"log"

	"github.com/example/ybMigration/internal/config"
)

// 演示新的配置使用方式
func main() {
	// 方式1: 直接加载默认配置
	fmt.Println("=== 方式1: 加载默认配置 ===")
	defaultConfig, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("加载默认配置失败: %v", err)
	}

	fmt.Printf("默认配置包含 %d 个规则\n", len(defaultConfig.GetRules()))

	// 方式2: 加载自定义配置文件
	fmt.Println("\n=== 方式2: 加载自定义配置 ===")
	customConfig, err := config.LoadConfig("configs/default.yaml")
	if err != nil {
		log.Fatalf("加载自定义配置失败: %v", err)
	}

	fmt.Printf("自定义配置包含 %d 个规则\n", len(customConfig.GetRules()))

	// 方式3: 按类别获取规则
	fmt.Println("\n=== 方式3: 按类别获取规则 ===")
	functionRules := defaultConfig.GetRulesByCategory("function")
	datatypeRules := defaultConfig.GetRulesByCategory("datatype")
	syntaxRules := defaultConfig.GetRulesByCategory("syntax")
	charsetRules := defaultConfig.GetRulesByCategory("charset")

	fmt.Printf("函数规则: %d 个\n", len(functionRules))
	fmt.Printf("数据类型规则: %d 个\n", len(datatypeRules))
	fmt.Printf("语法规则: %d 个\n", len(syntaxRules))
	fmt.Printf("字符集规则: %d 个\n", len(charsetRules))

	// 显示部分规则详情
	fmt.Println("\n=== 函数规则详情 ===")
	for i, rule := range functionRules {
		if i >= 3 { // 只显示前3个
			break
		}
		fmt.Printf("  %d. %s: %s -> %s\n", i+1, rule.Name, rule.When.Pattern, rule.Then.Target)
	}
}
