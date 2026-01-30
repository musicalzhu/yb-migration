# SQL 分析报告

## 摘要

- **来源**: c:\微云同步助手\1469866358\WorkBench\go\yb-migration\testdata\general_log_example.log
- **SQL 语句**:
  ```sql
  SELECT * FROM users;
UPDATE users SET name = 'test' WHERE id = 1;
SELECT IFNULL(orderid, 'N/A') FROM orders;

  ```
- **发现的问题数**: 1

## 发现的问题

### 问题 1: FunctionChecker

- **描述**: 函数 IFNULL: MySQL IFNULL 函数转换为标准 SQL COALESCE (建议: COALESCE)
- **自动修复**: 可用
  - **操作**: replace_function
  - **修复代码**:
    ```sql
    IFNULL -> COALESCE
    ```

---

