# 危险操作授权指南

**⚠️ 重要提醒**: 以下操作具有破坏性，执行前必须仔细确认并获得授权！

---

## 🔥 高危操作 (需要严格授权)

### 1. `git clean -fd` 
**危险等级**: 🔥🔥🔥🔥🔥 (极高)

**破坏性**: 删除所有未跟踪的文件和目录

**实际案例**:
```bash
# ❌ 错误操作 - 删除了重要文档
git clean -fd
# 结果: PROJECT_STATS.md, CODE_REVIEW_REPORT.md 等文件被删除
```

**安全替代方案**:
```bash
# ✅ 先查看将要删除的文件
git clean -n -fd

# ✅ 只删除特定类型的文件
git clean -fd -e "*.md" -e "*.txt"

# ✅ 或者手动删除确认的文件
rm -rf bin/ output-report/ temp/
```

**授权检查清单**:
- [ ] 确认没有未跟踪的重要文件
- [ ] 使用 `git status` 检查未跟踪文件
- [ ] 使用 `git clean -n` 预览删除操作
- [ ] 备份重要文件到安全位置

### 2. `git reset --hard HEAD~N`
**危险等级**: 🔥🔥🔥🔥 (高)

**破坏性**: 丢弃最近的 N 个提交

**安全替代方案**:
```bash
# ✅ 先创建备份分支
git branch backup-before-reset

# ✅ 使用软重置保留更改
git reset --soft HEAD~N

# ✅ 或者使用 revert 创建反向提交
git revert HEAD~N..HEAD
```

### 3. `rm -rf` 递归删除
**危险等级**: 🔥🔥🔥🔥 (高)

**破坏性**: 永久删除文件和目录

**安全替代方案**:
```bash
# ✅ 先移动到临时目录
mv target_dir /tmp/target_dir_backup

# ✅ 或者使用 trash 命令
trash target_dir

# ✅ 确认后再删除
ls -la target_dir  # 确认内容
rm -rf target_dir   # 删除
```

---

## ⚠️ 中危操作 (需要谨慎)

### 1. `git push --force`
**危险等级**: ⚠️⚠️⚠️ (中高)

**风险**: 覆盖远程分支历史

**安全替代方案**:
```bash
# ✅ 使用 --force-with-lease
git push --force-with-lease origin main

# ✅ 或者创建新分支
git checkout -b main-fixed
git push origin main-fixed
```

### 2. 大规模重构
**危险等级**: ⚠️⚠️ (中)

**风险**: 影响多个文件的稳定性

**安全替代方案**:
```bash
# ✅ 创建功能分支
git checkout -b refactor/feature-name

# ✅ 小步提交，频繁测试
git add .
git commit -m "refactor: step 1 - rename interface"

# ✅ 运行完整测试套件
go test ./...
```

---

## 📋 操作授权检查清单

### 执行前检查
- [ ] **备份重要文件**: 确认关键文件已备份
- [ ] **检查 Git 状态**: `git status` 查看未提交更改
- [ ] **预览操作**: 使用 `-n` 或 `--dry-run` 参数
- [ ] **确认目标路径**: 仔细检查文件路径和参数
- [ ] **通知团队成员**: 重要操作前通知相关人员

### 执行后验证
- [ ] **检查结果**: 确认操作结果符合预期
- [ ] **运行测试**: 执行完整测试套件
- [ ] **检查功能**: 验证核心功能正常
- [ ] **提交更改**: 及时提交成功的更改

---

## 🚨 事故案例记录

### 案例 1: git clean -fd 误删文档
**时间**: 2026-02-01  
**操作**: `git clean -fd`  
**后果**: 删除了 PROJECT_STATS.md 和 CODE_REVIEW_REPORT.md  
**教训**: 
- 重要文件要及时添加到 Git
- 使用 `git clean -n` 预览删除内容
- 谨慎使用 `-fd` 参数

### 案例 2: 大规模重构未测试
**时间**: 2026-01-30  
**操作**: 批量重命名接口  
**后果**: 多个测试失败，需要回滚  
**教训**:
- 大规模重构要分步进行
- 每步都要运行测试验证
- 使用功能分支隔离更改

---

## 🛡️ 安全操作最佳实践

### 1. Git 工作流安全
```bash
# ✅ 安全的工作流
git checkout -b feature/safe-operation
# 进行更改
git add .
git commit -m "feat: implement safe operation"
go test ./...  # 运行测试
git checkout main
git merge feature/safe-operation
git push origin main
```

### 2. 文件操作安全
```bash
# ✅ 安全的文件操作
# 1. 先查看
ls -la target/

# 2. 备份
cp -r target/ target_backup/

# 3. 操作
rm -rf target/

# 4. 验证
ls -la target_backup/
```

### 3. 批量操作安全
```bash
# ✅ 安全的批量操作
# 1. 生成操作列表
find . -name "*.tmp" > files_to_delete.txt

# 2. 检查列表
cat files_to_delete.txt

# 3. 执行删除
xargs rm -f < files_to_delete.txt
```

---

## 📞 紧急恢复方案

### 如果误删了重要文件
```bash
# 1. 检查 Git 历史
git log --name-status --all

# 2. 从历史恢复
git checkout <commit-hash> -- path/to/file

# 3. 或者从暂存区恢复
git checkout HEAD -- path/to/file
```

### 如果误提交了错误更改
```bash
# 1. 创建备份分支
git branch backup-mistake

# 2. 撤销提交
git reset --soft HEAD~1

# 3. 修复问题后重新提交
git add .
git commit -m "fix: correct the mistake"
```

---

## 🎯 授权原则

1. **最小权限原则**: 只执行必要的操作
2. **渐进式操作**: 小步快跑，频繁验证
3. **备份优先**: 任何破坏性操作前先备份
4. **团队协作**: 重要操作前与团队确认
5. **文档记录**: 记录重要操作和决策过程

---

**记住**: 谨慎操作，安全第一！任何不确定的操作都应该先咨询或有经验的人员指导。
