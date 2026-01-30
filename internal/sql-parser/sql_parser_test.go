package sqlparser

import (
	"testing"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseSQL_Basic 测试基本的SQL解析功能
func TestParseSQL_Basic(t *testing.T) {
	tests := []struct {
		name        string
		sql         string
		expectedLen int
		expectError bool
		expectNil   bool
	}{
		// 基本查询
		{
			name:        "简单SELECT语句",
			sql:         "SELECT * FROM users",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "带WHERE子句的SELECT",
			sql:         "SELECT * FROM users WHERE id = 1",
			expectedLen: 1,
			expectError: false,
		},
		// 数据操作
		{
			name:        "INSERT语句",
			sql:         "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "UPDATE语句",
			sql:         "UPDATE users SET name = 'Jane' WHERE id = 1",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "DELETE语句",
			sql:         "DELETE FROM users WHERE id = 1",
			expectedLen: 1,
			expectError: false,
		},
		// 表操作
		{
			name:        "CREATE TABLE语句",
			sql:         "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100))",
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "DROP TABLE语句",
			sql:         "DROP TABLE users",
			expectedLen: 1,
			expectError: false,
		},
		// 多语句
		{
			name:        "多条SQL语句",
			sql:         "SELECT * FROM users; INSERT INTO logs (action) VALUES ('test');",
			expectedLen: 2,
			expectError: false,
		},
		// 边界情况
		{
			name:        "空SQL",
			sql:         "",
			expectedLen: 0,
			expectError: false,
			expectNil:   true,
		},
		{
			name:        "只有空白字符",
			sql:         "   \t\n   ",
			expectedLen: 0,
			expectError: false,
			expectNil:   true,
		},
		// 错误情况
		{
			name:        "无效的SQL语法",
			sql:         "SELCT * FROM users", // SELECT拼写错误
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewSQLParser()
			stmts, err := parser.ParseSQL(tt.sql)

			if tt.expectError {
				assert.Error(t, err, "应该返回错误：%s", tt.sql)
				assert.Nil(t, stmts)
			} else {
				assert.NoError(t, err)
				if tt.expectNil {
					assert.Nil(t, stmts)
				} else {
					assert.NotNil(t, stmts)
					assert.Equal(t, tt.expectedLen, len(stmts))
				}
			}
		})
	}
}

// TestParseSQL_StatementTypes 测试解析不同类型的SQL语句
func TestParseSQL_StatementTypes(t *testing.T) {
	sql := "SELECT * FROM users; INSERT INTO users (name) VALUES ('test'); UPDATE users SET active = 1;"

	parser := NewSQLParser()
	stmts, err := parser.ParseSQL(sql)
	require.NoError(t, err)
	require.Len(t, stmts, 3)

	// 测试第一条语句是SELECT
	_, ok := stmts[0].(*ast.SelectStmt)
	assert.True(t, ok, "第一条语句应该是SELECT")

	// 测试第二条语句是INSERT
	_, ok = stmts[1].(*ast.InsertStmt)
	assert.True(t, ok, "第二条语句应该是INSERT")

	// 测试第三条语句是UPDATE
	_, ok = stmts[2].(*ast.UpdateStmt)
	assert.True(t, ok, "第三条语句应该是UPDATE")
}

// TestParseSQL_ComplexSQL 测试解析复杂的SQL语句
func TestParseSQL_ComplexSQL(t *testing.T) {
	tests := []struct {
		name string
		sql  string
	}{
		{
			name: "带约束的CREATE TABLE语句",
			sql: `CREATE TABLE orders (
				id INT PRIMARY KEY AUTO_INCREMENT,
				user_id INT NOT NULL,
				amount DECIMAL(10,2) DEFAULT 0.00,
				status ENUM('pending', 'completed', 'cancelled') DEFAULT 'pending',
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id)
			)`,
		},
		{
			name: "修改表结构",
			sql:  "ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE",
		},
		{
			name: "创建索引",
			sql:  "CREATE INDEX idx_users_email ON users(email)",
		},
		{
			name: "带GROUP BY和HAVING的复杂SELECT",
			sql:  "SELECT status, COUNT(*) as count FROM users GROUP BY status HAVING COUNT(*) > 10 ORDER BY count DESC",
		},
		{
			name: "插入多行数据",
			sql:  "INSERT INTO users (name, email) VALUES ('John', 'john@example.com'), ('Jane', 'jane@example.com')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewSQLParser()
			stmts, err := parser.ParseSQL(tt.sql)
			assert.NoError(t, err)
			assert.Len(t, stmts, 1)
			assert.NotNil(t, stmts[0])
		})
	}
}

// TestParseSQL_EdgeCases 测试边缘情况的SQL语句解析
func TestParseSQL_EdgeCases(t *testing.T) {
	t.Run("带注释的SQL", func(t *testing.T) {
		sql := `-- 这是一个注释
		SELECT * FROM users; /* 另一个注释 */ INSERT INTO logs (msg) VALUES ('test');`

		parser := NewSQLParser()
		stmts, err := parser.ParseSQL(sql)
		assert.NoError(t, err)
		assert.Len(t, stmts, 2)
	})
}

// TestParseSQL_ErrorHandling 测试错误处理的SQL语句解析
func TestParseSQL_ErrorHandling(t *testing.T) {
	tests := []struct {
		name string
		sql  string
	}{
		{
			name: "语法错误",
			sql:  "SELECT * FROM users WHERE",
		},
		{
			name: "不完整的关键字",
			sql:  "SELEC * FROM users",
		},
		{
			name: "无效的SQL语法",
			sql:  "SELECT * FROM table WHERE 1 = ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewSQLParser()
			stmts, err := parser.ParseSQL(tt.sql)
			assert.Error(t, err)
			assert.Nil(t, stmts)
		})
	}
}
