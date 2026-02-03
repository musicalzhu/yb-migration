package inputparser

import (
	"os"
	"strings"
	"testing"

	"github.com/example/ybMigration/internal/testutils"
)

func TestLogParser_Parse(t *testing.T) {
	testFile := testutils.MustGetTestDataPath("general_log_example.log")
	unsupportedFile := t.TempDir() + string(os.PathSeparator) + "unsupported.json"
	requireWriteFile(t, unsupportedFile, "{}")

	tests := []struct {
		name       string
		filePath   string
		wantSQL    string // 期望的SQL内容
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "正常日志文件",
			filePath: testFile,
			// 断言点：包含关键 SQL、每条语句以分号+换行结尾、语句数量稳定
			wantSQL: "",
			wantErr: false,
		},
		{
			name:       "不存在的文件",
			filePath:   testutils.MustGetTestDataPath("general_log_enot_existxample.log"),
			wantErr:    true,
			wantErrMsg: "不存在",
		},
		{
			name:       "不支持的扩展名",
			filePath:   unsupportedFile,
			wantErr:    true,
			wantErrMsg: "不支持的文件类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewGeneralLogFileParser()
			got, err := p.Parse(tt.filePath)

			// 检查错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.wantErrMsg != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrMsg)) {
					t.Fatalf("Parse() error = %v, want error containing %q", err, tt.wantErrMsg)
				}
				return
			}

			if tt.wantSQL != "" {
				if got != tt.wantSQL {
					t.Errorf("SQL content mismatch:\ngot:  %q\nwant: %q", got, tt.wantSQL)
				}
				return
			}

			// 正常日志文件：关键片段断言（避免对整段字符串全等的脆弱性）
			if strings.TrimSpace(got) == "" {
				t.Fatalf("Parse() returned empty SQL content")
			}

			assertContainsAll(t, got,
				"SELECT * FROM users;\n",
				"UPDATE users SET name = 'test' WHERE id = 1;\n",
				"SELECT IFNULL(orderid, 'N/A') FROM orders;\n",
			)

			// 应该只包含 3 条 Query 语句（testdata 中共 3 行 Query）
			assertEqual(t, 3, strings.Count(got, ";\n"))

			// 非 Query 行不应出现在 SQL 输出中
			if strings.Contains(got, "Connect") || strings.Contains(got, "Quit") {
				t.Fatalf("Parse() output should not include non-query lines: %q", got)
			}

			// 非 Query 行会被记录到 nonStandardLines（Connect/Quit）
			nonStandard := p.GetNonStandardLines()
			if len(nonStandard) < 2 {
				t.Fatalf("expected nonStandardLines to record non-query lines, got: %v", nonStandard)
			}
		})
	}
}

// assertEqual 断言两个值相等（泛型版本）
// 参数:
//   - t: 测试实例
//   - want: 期望值
//   - got: 实际值
func assertEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// assertContainsAll 断言字符串包含所有指定的子字符串
// 参数:
//   - t: 测试实例
//   - s: 要检查的字符串
//   - subs: 必须包含的子字符串列表
func assertContainsAll(t *testing.T, s string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			t.Fatalf("expected output to contain %q, got: %q", sub, s)
		}
	}
}

// requireWriteFile 写入测试文件，失败时终止测试
// 参数:
//   - t: 测试实例
//   - path: 文件路径
//   - content: 文件内容
//
// 注意事项:
//   - 文件权限设置为 0644
//   - 写入失败时调用 t.Fatalf 终止测试
func requireWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("write temp file failed: %v", err)
	}
}
