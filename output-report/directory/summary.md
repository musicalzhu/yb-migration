# SQL 分析报告

## 摘要

- **来源**: c:\微云同步助手\1469866358\WorkBench\go\yb-migration\testdata
- **发现的问题数**: 3

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

### 问题 2: FunctionChecker

- **描述**: 函数 GROUP_CONCAT: MySQL GROUP_CONCAT 函数转换为标准 SQL STRING_AGG (建议: STRING_AGG)
- **自动修复**: 可用
  - **操作**: replace_function
  - **修复代码**:
    ```sql
    GROUP_CONCAT -> STRING_AGG
    ```

---

### 问题 3: SyntaxChecker

- **描述**: 语法 AUTO_INCREMENT: MySQL AUTO_INCREMENT 转换为 PostgreSQL SERIAL (建议: SERIAL)
- **自动修复**: 可用
  - **操作**: replace_constraint
  - **修复代码**:
    ```sql
    AUTO_INCREMENT -> SERIAL
    ```

---

