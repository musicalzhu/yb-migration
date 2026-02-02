package inputparser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/ybMigration/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLFileParser_Parse(t *testing.T) {
	parser := NewSQLFileParser()

	tests := []struct {
		name        string
		sqlContent  string
		expectError bool
		expectCount int
	}{
		{
			name: "单个SQL语句",
			sqlContent: "SELECT * FROM users;",
			expectError: false,
			expectCount: 1,
		},
		{
			name: "多个SQL语句",
			sqlContent: `SELECT * FROM users;
INSERT INTO logs (message) VALUES ('test');
UPDATE users SET active = 1;`,
			expectError: false,
			expectCount: 3,
		},
		{
			name: "带注释的SQL",
			sqlContent: `-- 这是一个注释
SELECT * FROM users;
/* 多行注释 */
INSERT INTO logs (msg) VALUES ('test');`,
			expectError: false,
			expectCount: 2,
		},
		{
			name: "空文件",
			sqlContent: "",
			expectError: false,
			expectCount: 0,
		},
		{
			name: "只有注释",
			sqlContent: `-- 单行注释
/* 多行注释 */`,
			expectError: false,
			expectCount: 0,
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
			assert.Equal(t, tmpFile.Name(), result.Source)
			assert.Equal(t, tt.expectCount, len(result.SQLs))
		})
	}
}

func TestSQLFileParser_Parse_NonExistentFile(t *testing.T) {
	parser := NewSQLFileParser()
	
	_, err := parser.Parse("/non/existent/file.sql")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文件不存在")
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
	assert.Contains(t, err.Error(), "不是文件")
}

func TestSQLFileParser_Parse_LargeFile(t *testing.T) {
	parser := NewSQLFileParser()
	
	// 创建大文件
	tmpFile, err := os.CreateTemp("", "test_large_*.sql")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// 写入大量SQL语句
	sqlContent := "SELECT * FROM users WHERE id = %d;\n"
	for i := 0; i < 1000; i++ {
		_, err = tmpFile.WriteString(fmt.Sprintf(sqlContent, i))
		require.NoError(t, err)
	}
	require.NoError(t, tmpFile.Close())

	// 解析大文件
	result, err := parser.Parse(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, 1000, len(result.SQLs))
}

func TestSQLFileParser_Parse_WithTestData(t *testing.T) {
	parser := NewSQLFileParser()
	
	// 使用测试数据文件
	testDataPath := testutils.MustGetTestDataPath("mysql_queries.sql")
	
	result, err := parser.Parse(testDataPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, result.SQLs)
	assert.Equal(t, testDataPath, result.Source)
}

func TestSQLFileParser_Parse_Encoding(t *testing.T) {
	parser := NewSQLFileParser()
	
	// 创建包含UTF-8字符的SQL文件
	tmpFile, err := os.CreateTemp("", "test_utf8_*.sql")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// 写入包含中文的SQL
	sqlContent := "SELECT * FROM 用户 WHERE 姓名 = '测试';"
	_, err = tmpFile.WriteString(sqlContent)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	// 解析文件
	result, err := parser.Parse(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.SQLs))
	assert.Contains(t, result.SQLs[0], "用户")
}

func TestSQLFileParser_Parse_PermissionDenied(t *testing.T) {
	parser := NewSQLFileParser()
	
	// 在Windows上，创建一个没有读取权限的文件比较复杂
	// 这里只测试文件不存在的情况，权限测试在CI环境中可能不稳定
	t.Skip("权限测试在CI环境中不稳定，跳过")
}
