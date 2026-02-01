package checker

import (
	"fmt"

	"github.com/example/ybMigration/internal/config"
	"github.com/example/ybMigration/internal/model"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/parser/types"
)

// DataTypeChecker 数据类型检查器实现
// 检查SQL数据类型兼容性问题
// 支持从default.yaml配置文件加载规则，实现MySQL到目标数据库的语法转换
type DataTypeChecker struct {
	*RuleChecker
}

// NewDataTypeChecker 创建数据类型检查器
// 返回:
//   - *DataTypeChecker: 初始化后的数据类型检查器实例
//   - error: 错误信息
func NewDataTypeChecker(cfg *config.Config) (*DataTypeChecker, error) {
	ruleChecker, err := newRuleChecker("DataTypeChecker", "datatype", cfg)
	if err != nil {
		return nil, fmt.Errorf("创建数据类型检查器失败: %w", err)
	}
	return &DataTypeChecker{
		RuleChecker: ruleChecker,
	}, nil
}

// Name 返回检查器名称
func (d *DataTypeChecker) Name() string {
	return "DataTypeChecker"
}

// Inspect 实现 Checker 接口，处理 AST 节点
// 检查数据类型兼容性并执行转换
func (d *DataTypeChecker) Inspect(n ast.Node) (w ast.Node, skipChildren bool) {
	switch node := n.(type) {
	case *ast.ColumnDef:
		// 检查并转换列定义中的数据类型
		return d.checkColumnType(node)

	case *ast.AlterTableStmt:
		// 检查并转换修改表结构中的数据类型变更
		return d.checkAlterTable(node)
	}
	return n, false
}

// checkColumnType 检查列数据类型兼容性并执行转换
// 实现MySQL数据类型到目标数据库的转换
// 参数:
//   - node: 列定义节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (d *DataTypeChecker) checkColumnType(node *ast.ColumnDef) (ast.Node, bool) {
	if node == nil || node.Tp == nil {
		return node, false
	}

	// 获取数据类型名称
	typeName := d.extractTypeNameFromFieldType(node.Tp)
	if typeName == "" {
		return node, false
	}

	// 检查是否有匹配的规则
	rules := d.GetRules()
	rule, hasRule := rules[typeName]
	if !hasRule {
		return node, false
	}

	// 生成兼容性问题
	d.AddIssue(model.Issue{
		Checker: "DataTypeChecker",
		Message: fmt.Sprintf("数据类型 %s: %s (建议: %s)", typeName, rule.Description, rule.Then.Target),
		AutoFix: model.AutoFix{
			Available: true,
			Action:    rule.Then.Action,
			Code:      fmt.Sprintf("%s -> %s", typeName, rule.Then.Target),
		},
	})

	// 执行AST转换
	transformedNode := d.ApplyTransformation(node, rule)
	return transformedNode, true
}

// extractTypeNameFromFieldType 从 FieldType 提取类型名称。
// 使用 TiDB 的 GetType() 方法直接获取类型常量，并转换为字符串表示。
// 参数:
//   - tp: FieldType 实例，不能为 nil
//
// 返回:
//   - string: 类型名称（如 "INT", "VARCHAR" 等），如果类型未知或未指定则返回空字符串
func (d *DataTypeChecker) extractTypeNameFromFieldType(tp *types.FieldType) string {
	if tp == nil {
		return ""
	}

	// 使用 GetType() 方法直接获取类型常量
	switch tp.GetType() {
	case mysql.TypeTiny:
		return "TINYINT"
	case mysql.TypeShort:
		return "SMALLINT"
	case mysql.TypeLong:
		return "INT"
	case mysql.TypeInt24:
		return "MEDIUMINT"
	case mysql.TypeLonglong:
		return "BIGINT"
	case mysql.TypeFloat:
		return "FLOAT"
	case mysql.TypeDouble:
		return "DOUBLE"
	case mysql.TypeNewDecimal:
		return "DECIMAL"
	case mysql.TypeDate:
		return "DATE"
	case mysql.TypeDatetime:
		return "DATETIME"
	case mysql.TypeTimestamp:
		return "TIMESTAMP"
	case mysql.TypeDuration:
		return "TIME"
	case mysql.TypeYear:
		return "YEAR"
	case mysql.TypeVarchar:
		return "VARCHAR"
	case mysql.TypeString:
		return "CHAR"
	case mysql.TypeVarString:
		return "VARCHAR"
	case mysql.TypeBlob:
		return "BLOB"
	case mysql.TypeTinyBlob:
		return "TINYBLOB"
	case mysql.TypeMediumBlob:
		return "MEDIUMBLOB"
	case mysql.TypeLongBlob:
		return "LONGBLOB"
	case mysql.TypeJSON:
		return "JSON"
	case mysql.TypeEnum:
		return "ENUM"
	case mysql.TypeSet:
		return "SET"
	case mysql.TypeBit:
		return "BIT"
	case mysql.TypeGeometry:
		return "GEOMETRY"
	case mysql.TypeUnspecified:
		return ""
	default:
		// 返回空字符串而不是 panic，调用者应检查返回值
		// 对于未知类型，返回空字符串表示无法识别该类型
		// 调用者应检查返回值是否为空，以判断类型是否有效
		return ""
	}
}

// checkAlterTable 检查ALTER TABLE语句中的数据类型变更
// 处理表结构变更中的数据类型兼容性
// 参数:
//   - node: ALTER TABLE语句节点
//
// 返回值:
//   - ast.Node: 转换后的节点
//   - bool: 是否有转换发生
func (d *DataTypeChecker) checkAlterTable(node *ast.AlterTableStmt) (ast.Node, bool) {
	hasTransform := false

	// 遍历ALTER TABLE中的所有变更项
	for _, spec := range node.Specs {
		alterSpec := spec
		switch alterSpec.Tp {
		case ast.AlterTableAddColumns:
			// 检查新增列的数据类型
			if alterSpec.NewColumns != nil {
				for _, col := range alterSpec.NewColumns {
					if transformedNode, transformed := d.checkColumnType(col); transformed {
						*col = *transformedNode.(*ast.ColumnDef)
						hasTransform = true
					}
				}
			}

		case ast.AlterTableModifyColumn:
			// 检查修改列的数据类型
			if len(alterSpec.NewColumns) > 0 {
				if transformedNode, transformed := d.checkColumnType(alterSpec.NewColumns[0]); transformed {
					*alterSpec.NewColumns[0] = *transformedNode.(*ast.ColumnDef)
					hasTransform = true
				}
			}

		case ast.AlterTableChangeColumn:
			// 检查变更列的数据类型
			if len(alterSpec.NewColumns) > 0 {
				if transformedNode, transformed := d.checkColumnType(alterSpec.NewColumns[0]); transformed {
					*alterSpec.NewColumns[0] = *transformedNode.(*ast.ColumnDef)
					hasTransform = true
				}
			}
		}
	}

	return node, hasTransform
}
