package inputparser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLFileParser_Parse(t *testing.T) {
	parser := NewSQLFileParser()

	tests := []struct {
		name           string
		sqlContent     string
		expectError    bool
		expectContains []string
	}{
		{
			name:           "单个SQL语句",
			sqlContent:     "SELECT * FROM users;",
			expectError:    false,
			expectContains: []string{"SELECT * FROM users"},
		},
		{
			name: "多个SQL语句",
			sqlContent: `SELECT * FROM users;
INSERT INTO logs (message) VALUES ('test');
UPDATE users SET active = 1;`,
			expectError:    false,
			expectContains: []string{"SELECT * FROM users", "INSERT INTO logs", "UPDATE users SET active = 1"},
		},
		{
			name: "带注释的SQL",
			sqlContent: `-- 这是一个注释
SELECT * FROM users;
/* 多行注释 */
INSERT INTO logs (msg) VALUES ('test');`,
			expectError:    false,
			expectContains: []string{"SELECT * FROM users", "INSERT INTO logs"},
		},
		{
			name:           "空文件",
			sqlContent:     "",
			expectError:    false,
			expectContains: []string{}, // 空文件返回空字符串
		},
		{
			name: "只有注释",
			sqlContent: `-- 单行注释
/* 多行注释 */`,
			expectError:    false,
			expectContains: []string{"-- 单行注释", "/* 多行注释 */"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时文件
			tmpFile, err := os.CreateTemp("", "test_*.sql")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			// 写入测试内容
			_, err = tmpFile.WriteString(tt.sqlContent)
			require.NoError(t, err)
			require.NoError(t, tmpFile.Close())

			// 解析文件
			result, err := parser.Parse(tmpFile.Name())

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// 空文件应该返回空字符串，这是正常的
			if len(tt.expectContains) == 0 {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
			}

			// 检查是否包含期望的内容
			for _, expected := range tt.expectContains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

func TestSQLFileParser_Parse_NonExistentFile(t *testing.T) {
	parser := NewSQLFileParser()

	_, err := parser.Parse("/non/existent/file.sql")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不存在")
}

func TestSQLFileParser_Parse_InvalidExtension(t *testing.T) {
	parser := NewSQLFileParser()

	// 创建一个非.sql文件
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	require.NoError(t, tmpFile.Close())

	_, err = parser.Parse(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的文件类型")
}

func TestSQLFileParser_Parse_Directory(t *testing.T) {
	parser := NewSQLFileParser()

	// 尝试解析目录
	_, err := parser.Parse(os.TempDir())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持目录")
}
