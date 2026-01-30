package inputparser

import (
	"strings"
	"testing"

	"github.com/example/ybMigration/internal/testutils"
)

func TestLogParser_ParseFile(t *testing.T) {
	testFile := testutils.MustGetTestDataPath("general_log_example.log")

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
			// 期望的SQL内容，包含所有查询语句，用分号分隔
			wantSQL: "SELECT * FROM users;\nUPDATE users SET name = 'test' WHERE id = 1;\nSELECT IFNULL(orderid, 'N/A') FROM orders;\n",
			wantErr: false,
		},
		{
			name:       "不存在的文件",
			filePath:   testutils.MustGetTestDataPath("general_log_enot_existxample.log"),
			wantErr:    true,
			wantErrMsg: "打开文件",
		},
		{
			name:       "不支持的扩展名",
			filePath:   testutils.MustGetTestDataPath("merged.report.json"),
			wantErr:    true,
			wantErrMsg: "不支持的文件类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewGeneralLogFileParser()
			got, err := p.ParseFile(tt.filePath)

			// 检查错误
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.wantErrMsg != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrMsg)) {
					t.Fatalf("ParseFile() error = %v, want error containing %q", err, tt.wantErrMsg)
				}
				return
			}

			// 检查返回的SQL内容
			if got != tt.wantSQL {
				t.Errorf("SQL content mismatch:\ngot:  %q\nwant: %q", got, tt.wantSQL)
			}
		})
	}
}
