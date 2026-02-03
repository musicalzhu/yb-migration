# GitLab 社区版 CI/CD 配置指南

## 🏢 GitLab 社区版管理支持

### 🔧 GitLab 社区版优化

#### 1. 项目配置
```yaml
# .gitlab-ci.yml 中的 GitLab 社区版配置
variables:
  # GitLab 社区版配置
  GIT_DEPTH: 0                    # 完整克隆，用于质量分析
  GIT_STRATEGY: clone             # 完整克隆策略
  CACHE_KEY_PREFIX: "yb-migration"
  
  # 社区版配置
  DOCKER_REGISTRY: "registry.gitlab.com"
  DOCKER_IMAGE: "${DOCKER_REGISTRY}/yb-migration/${APP_NAME}"
  
  # 通知配置（社区版支持）
  NOTIFICATION_WEBHOOK: "${WEBHOOK_URL}"
  SLACK_WEBHOOK: "${SLACK_URL}"
```

#### 2. 分支策略
```yaml
# 支持的分支模式
only:
  - main                         # 主分支
  - develop                      # 开发分支
  - merge_requests              # 合并请求
  - /^release\/.*$/             # release 分支
  - /^hotfix\/.*$/              # hotfix 分支
  - tags                        # 标签推送
```

### 🎯 质量门禁流程

#### 阶段设计
```yaml
stages:
  - prepare                      # 环境准备
  - quality                      # 质量检查
  - test                         # 测试阶段
  - security                     # 安全扫描
  - build                        # 构建阶段
  - deploy                       # 部署阶段
  - notify                       # 通知阶段
```

#### 依赖关系
```yaml
# 质量检查失败会阻塞后续流程
quality-check:
  allow_failure: false           # 必须通过

# 安全扫描失败仅警告
security-scan:
  allow_failure: true            # 警告但不阻塞

# 测试依赖质量检查
unit-test:
  dependencies:
    - quality-check
    - coverage-gate
```

### 📊 报告集成

#### GitLab 原生报告
```yaml
# JUnit 测试报告
reports:
  junit: test-report.xml

# 覆盖率报告
coverage: '/total:.*?(\d+\.\d+)%/'

# 产物管理
artifacts:
  reports:
    junit: quality-checkstyle.xml
```

#### 质量指标收集
```yaml
# 生成质量指标
script:
  - |
    cat > quality-metrics.json << EOF
    {
      "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
      "commit": "${CI_COMMIT_SHORT_SHA}",
      "branch": "${CI_COMMIT_REF_NAME}",
      "project": "${CI_PROJECT_PATH}",
      "pipeline": "${CI_PIPELINE_ID}",
      "job": "${CI_JOB_ID}"
    }
    EOF
```

### 🔒 安全配置

#### 社区版安全策略
```yaml
# 敏感信息检测
script:
  - |
    if grep -r -i "password\|secret\|key\|token" --include="*.go" --exclude-dir=.git .; then
      echo "::warning::发现可能的敏感信息泄露"
    fi

# 依赖安全扫描（社区版支持）
script:
  - go list -json -m all | nancy sleuth

# 社区版安全扫描
include:
  - template: Security/License-Scanning.gitlab-ci.yml
  - template: Security/SAST.gitlab-ci.yml
  - template: Security/Secret-Detection.gitlab-ci.yml
```

#### 权限控制
```yaml
# 生产环境部署权限
deploy-production:
  only:
    - tags
  when: manual                 # 手动确认
  environment:
    name: production
    url: https://example.com
```

### 🚀 部署配置

#### 多环境支持
```yaml
# 测试环境
deploy-staging:
  environment:
    name: staging
    url: https://staging.example.com
  only:
    - main
    - develop
  when: manual

# 生产环境
deploy-production:
  environment:
    name: production
    url: https://example.com
  only:
    - tags
  when: manual
```

#### Docker 集成
```yaml
docker-build:
  image: docker:latest
  services:
    - docker:dind
  script:
    - docker build -t ${DOCKER_IMAGE}:${CI_COMMIT_SHORT_SHA} .
    - docker tag ${DOCKER_IMAGE}:${CI_COMMIT_SHORT_SHA} ${DOCKER_IMAGE}:latest
```

### 📢 通知集成

#### Slack 通知
```yaml
notify:
  script:
    - |
      if [ -n "$SLACK_WEBHOOK" ]; then
        curl -X POST -H 'Content-type: application/json' \
          --data "{\"text\":\"🚀 Pipeline 完成: ${CI_PROJECT_PATH}\"}" \
          "$SLACK_WEBHOOK"
      fi
```

#### GitLab 内置通知
```yaml
# 使用 GitLab 内置通知
variables:
  NOTIFICATION_WEBHOOK: "${WEBHOOK_URL}"
```

### 🏷️ 标签策略

#### 版本管理
```yaml
# 自动版本标签
build:
  script:
    - go build -ldflags="-X main.Version=${CI_COMMIT_TAG:-${CI_COMMIT_SHORT_SHA}}" ./cmd

# 发布管理
release:
  only:
    - tags
  when: manual
```

### 📈 性能监控

#### 基准测试
```yaml
benchmark:
  stage: test
  script:
    - go test -bench=. -benchmem -run=^$$ ./... > benchmark.txt
  artifacts:
    paths:
      - benchmark.txt
```

#### 构建优化
```yaml
# 缓存策略
cache:
  paths:
    - .cache/
    - .gocache/
    - vendor/
  key: "${CACHE_KEY_PREFIX}-${CI_COMMIT_REF_SLUG}-${CI_COMMIT_SHORT_SHA}"
```

### 🔧 GitLab CI/CD 变量

#### 必需变量
```bash
# 在 GitLab 项目设置中配置
WEBHOOK_URL="https://your-webhook-url"
SLACK_URL="https://hooks.slack.com/your-slack-webhook"
DOCKER_REGISTRY="registry.gitlab.com"  # 社区版使用公共 registry
```

#### 可选变量
```bash
# 性能调优
GO_VERSION="1.25.1"
CGO_ENABLED="0"
GOPROXY="https://goproxy.cn,direct"

# 社区版特定
CI_RUNNER_TAGS="docker,linux,quality,security,build,deploy"
```

### 📋 使用指南

#### 1. 项目设置
1. 在 GitLab 项目中设置 CI/CD 变量
2. 配置 Runner 标签 (`docker`, `linux`, `quality`, `security`, `build`, `deploy`)
3. 设置分支保护规则

#### 2. 本地开发
```bash
# 安装 GitLab Runner (本地测试)
curl -L https://packages.gitlab.com/install/repositories/runner/gitlab-runner/script.rpm.sh | sudo bash
sudo yum install gitlab-runner

# 注册 Runner (社区版)
sudo gitlab-runner register \
  --url "https://gitlab.com/" \
  --registration-token "YOUR_TOKEN" \
  --description "community-runner" \
  --tag-list "docker,linux,quality,security,build,deploy" \
  --executor "docker"
```

#### 3. 质量检查
```bash
# 本地运行质量检查
make quality-check

# 提交前检查
make pre-commit
```

### 🎯 最佳实践

#### 1. 分支命名
- `main` - 生产分支
- `develop` - 开发分支  
- `feature/xxx` - 功能分支
- `release/xxx` - 发布分支
- `hotfix/xxx` - 热修复分支

#### 2. 提交规范
- 使用语义化提交消息
- 关联相关 Issue
- 添加测试覆盖

#### 3. 版本管理
- 使用语义化版本号
- 自动生成 Change Log
- 标签触发发布

### 🚨 故障排除

#### 常见问题
1. **Runner 权限**: 确保社区版 Runner 有足够权限
2. **缓存问题**: 清理缓存或更新缓存键
3. **网络问题**: 配置代理或镜像
4. **依赖问题**: 使用 vendor 模式
5. **社区版限制**: 注意社区版与企业版的功能差异

#### 调试技巧
```yaml
# 启用调试模式
variables:
  CI_DEBUG_TRACE: "true"

# 保存调试信息
artifacts:
  when: always
  paths:
    - "*.log"
    - "debug/"
```

---

## 📞 社区版支持

如有问题，请联系：
- GitLab 社区版文档: https://docs.gitlab.com/ee/
- GitLab 社区论坛: https://forum.gitlab.com/
- 项目 Issues: https://gitlab.com/your-project/-/issues
