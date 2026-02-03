#!/bin/bash

# GitLab CI/CD 快速设置脚本
# 用于内部 GitLab 项目的自动化配置

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 GitLab CI/CD 快速设置${NC}"
echo "=================================="

# 检查是否在项目根目录
if [ ! -f "go.mod" ] || [ ! -f ".gitlab-ci.yml" ]; then
    echo -e "${RED}❌ 错误: 请在包含 go.mod 和 .gitlab-ci.yml 的项目根目录运行此脚本${NC}"
    exit 1
fi

# 项目配置
PROJECT_NAME=${1:-"yb-migration"}
GITLAB_URL=${2:-"gitlab.company.com"}
DOCKER_REGISTRY=${3:-"registry.gitlab.company.com"}

echo -e "${BLUE}📋 项目配置:${NC}"
echo "项目名称: ${PROJECT_NAME}"
echo "GitLab 地址: ${GITLAB_URL}"
echo "Docker Registry: ${DOCKER_REGISTRY}"
echo ""

# 创建 GitLab CI/CD 配置目录
mkdir -p .gitlab ci templates

# 创建 GitLab CI/CD 模板
echo -e "${BLUE}📝 创建 GitLab CI/CD 模板...${NC}"

cat > .gitlab/ci-variables.yml << EOF
# GitLab CI/CD 变量配置模板
# 在 GitLab 项目 Settings > CI/CD > Variables 中添加以下变量

# 必需变量
WEBHOOK_URL=https://your-webhook-url.com/api/notify
SLACK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK

# Docker 配置
DOCKER_REGISTRY=${DOCKER_REGISTRY}
DOCKER_USERNAME=gitlab-ci-token
DOCKER_PASSWORD=\${CI_JOB_TOKEN}

# 环境配置
STAGING_URL=https://staging.example.com
PRODUCTION_URL=https://example.com

# 通知配置
NOTIFICATION_EMAIL=devops@company.com
NOTIFICATION_CHANNEL=#devops-alerts
EOF

# 创建 Docker Compose 开发环境
cat > docker-compose.gitlab.yml << EOF
version: '3.8'

services:
  yb-migration:
    build:
      context: .
      dockerfile: Dockerfile.gitlab
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=debug
      - PORT=8080
    volumes:
      - ./configs:/app/configs
      - ./output-report:/app/output-report
    restart: unless-stopped

  # 可选: 添加数据库服务用于测试
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: yb_migration_test
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped

volumes:
  mysql_data:
EOF

# 创建 GitLab Runner 配置
cat > gitlab-runner-config.toml << EOF
# GitLab Runner 配置模板
# 保存为 /etc/gitlab-runner/config.toml

concurrent = 4
check_interval = 0

[[runners]]
  name = "docker-runner"
  url = "https://${GITLAB_URL}/"
  token = "YOUR_RUNNER_TOKEN"
  executor = "docker"
  [runners.docker]
    tls_verify = false
    image = "golang:1.25.1"
    privileged = false
    disable_entrypoint_overwrite = false
    oom_kill_disable = false
    disable_cache = false
    volumes = ["/cache"]
    shm_size = 0
  [runners.cache]
    [runners.cache.s3]
    [runners.cache.gcs]
    [runners.cache.azure]
EOF

# 创建部署脚本
cat > scripts/gitlab-deploy.sh << EOF
#!/bin/bash

# GitLab 部署脚本
# 用于 GitLab CI/CD 自动部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 环境变量
ENVIRONMENT=\${1:-"staging"}
VERSION=\${2:-\${CI_COMMIT_SHORT_SHA}}
DOCKER_IMAGE=\${3:-"\${DOCKER_REGISTRY}/${PROJECT_NAME}"}

echo -e "\${BLUE}🚀 部署到 \${ENVIRONMENT} 环境\${NC}"
echo "版本: \${VERSION}"
echo "镜像: \${DOCKER_IMAGE}"
echo ""

# 检查环境
case \${ENVIRONMENT} in
  "staging")
    DEPLOY_URL="\${STAGING_URL}"
    ;;
  "production")
    DEPLOY_URL="\${PRODUCTION_URL}"
    ;;
  *)
    echo -e "\${RED}❌ 不支持的环境: \${ENVIRONMENT}\${NC}"
    exit 1
    ;;
esac

# 部署逻辑
echo -e "\${BLUE}📦 拉取镜像...\${NC}"
docker pull \${DOCKER_IMAGE}:\${VERSION}

echo -e "\${BLUE}🚀 启动容器...\${NC}"
docker run -d \\
  --name \${PROJECT_NAME}-\${ENVIRONMENT} \\
  --restart unless-stopped \\
  -p 8080:8080 \\
  -e GIN_MODE=release \\
  -e PORT=8080 \\
  -v /app/configs:/app/configs \\
  -v /app/output-report:/app/output-report \\
  \${DOCKER_IMAGE}:\${VERSION}

echo -e "\${BLUE}🔍 健康检查...\${NC}"
sleep 10

if curl -f http://localhost:8080/health; then
    echo -e "\${GREEN}✅ 部署成功\${NC}"
    echo -e "\${GREEN}🌐 访问地址: \${DEPLOY_URL}\${NC}"
else
    echo -e "\${RED}❌ 健康检查失败\${NC}"
    docker logs \${PROJECT_NAME}-\${ENVIRONMENT}
    exit 1
fi

echo -e "\${GREEN}🎉 部署完成\${NC}"
EOF

chmod +x scripts/gitlab-deploy.sh

# 创建 GitLab CI/CD 变量设置脚本
cat > scripts/setup-gitlab-variables.sh << EOF
#!/bin/bash

# GitLab CI/CD 变量设置脚本
# 使用 GitLab API 自动设置项目变量

set -e

# 配置
GITLAB_URL="\${GITLAB_URL:-${GITLAB_URL}}"
PROJECT_ID="\${PROJECT_ID}"
PRIVATE_TOKEN="\${GITLAB_PRIVATE_TOKEN}"

# 检查必需参数
if [ -z "\$GITLAB_URL" ] || [ -z "\$PROJECT_ID" ] || [ -z "\$PRIVATE_TOKEN" ]; then
    echo "❌ 请设置以下环境变量:"
    echo "  GITLAB_URL: GitLab 实例地址"
    echo "  PROJECT_ID: GitLab 项目 ID"
    echo "  GITLAB_PRIVATE_TOKEN: GitLab API Token"
    exit 1
fi

echo "🔧 设置 GitLab CI/CD 变量..."

# 变量列表
variables=(
    "WEBHOOK_URL:https://your-webhook-url.com/api/notify"
    "SLACK_URL:https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
    "DOCKER_REGISTRY:${DOCKER_REGISTRY}"
    "STAGING_URL:https://staging.example.com"
    "PRODUCTION_URL:https://example.com"
    "NOTIFICATION_EMAIL:devops@company.com"
)

# 设置变量
for var in "\${variables[@]}"; do
    key="\${var%%:*}"
    value="\${var#*:}"
    
    echo "设置变量: \$key"
    
    curl --request PUT \\
        --url "\${GITLAB_URL}/api/v4/projects/\${PROJECT_ID}/variables/\$key" \\
        --header "PRIVATE-TOKEN: \${PRIVATE_TOKEN}" \\
        --header "Content-Type: application/json" \\
        --data "{\"value\": \"\$value\", \"protected\": false, \"masked\": false}" || \\
    curl --request POST \\
        --url "\${GITLAB_URL}/api/v4/projects/\${PROJECT_ID}/variables" \\
        --header "PRIVATE-TOKEN: \${PRIVATE_TOKEN}" \\
        --header "Content-Type: application/json" \\
        --data "{\"key\": \"\$key\", \"value\": \"\$value\", \"protected\": false, \"masked\": false}"
done

echo "✅ GitLab CI/CD 变量设置完成"
EOF

chmod +x scripts/setup-gitlab-variables.sh

# 创建 GitLab Runner 注册脚本
cat > scripts/register-gitlab-runner.sh << EOF
#!/bin/bash

# GitLab Runner 注册脚本

set -e

# 配置
GITLAB_URL="\${GITLAB_URL:-${GITLAB_URL}}"
RUNNER_TOKEN="\${RUNNER_TOKEN}"
RUNNER_NAME="\${RUNNER_NAME:-docker-runner}"
RUNNER_TAGS="\${RUNNER_TAGS:-docker,linux,quality,security,build,deploy}"

echo "🔧 注册 GitLab Runner..."

# 注册 Runner
sudo gitlab-runner register \\
    --non-interactive \\
    --url "\${GITLAB_URL}" \\
    --registration-token "\${RUNNER_TOKEN}" \\
    --name "\${RUNNER_NAME}" \\
    --tag-list "\${RUNNER_TAGS}" \\
    --run-untagged="false" \\
    --docker-privileged="true" \\
    --docker-image="golang:1.25.1" \\
    --docker-pull-policy="if-not-present" \\
    --executor "docker"

echo "✅ GitLab Runner 注册完成"

# 启动 Runner
sudo gitlab-runner start

echo "🚀 GitLab Runner 已启动"
EOF

chmod +x scripts/register-gitlab-runner.sh

# 创建项目 README 更新
cat > README-GitLab.md << EOF
# 🚀 YB Migration - GitLab 企业版

## 📋 项目概述

YB Migration 是一个 MySQL 到 YB 数据库迁移兼容性分析工具，专为内部 GitLab 环境优化。

## 🏢 GitLab 集成

### CI/CD 流水线
- **质量门禁**: 代码质量、安全扫描、格式检查
- **自动化测试**: 单元测试、集成测试、性能测试
- **多环境部署**: 测试环境、生产环境
- **通知集成**: Slack、邮件、Webhook

### 分支策略
- \`main\`: 生产分支
- \`develop\`: 开发分支
- \`feature/*\`: 功能分支
- \`release/*\`: 发布分支
- \`hotfix/*\`: 热修复分支

## 🚀 快速开始

### 1. 本地开发
\`\`\`bash
# 克隆项目
git clone https://${GITLAB_URL}/teams/${PROJECT_NAME}.git
cd ${PROJECT_NAME}

# 安装依赖
./scripts/setup-dev.sh

# 运行测试
make test

# 本地构建
make build
\`\`\`

### 2. GitLab CI/CD
\`\`\`bash
# 设置 GitLab 变量
./scripts/setup-gitlab-variables.sh

# 注册 Runner
./scripts/register-gitlab-runner.sh
\`\`\`

### 3. Docker 开发
\`\`\`bash
# 启动开发环境
docker-compose -f docker-compose.gitlab.yml up -d

# 查看日志
docker-compose -f docker-compose.gitlab.yml logs -f
\`\`\`

## 📊 质量指标

- **代码覆盖率**: ≥ 60%
- **圈复杂度**: ≤ 12
- **函数长度**: ≤ 80 行
- **安全漏洞**: 0 高危

## 🔒 安全配置

- 敏感信息检测
- 依赖漏洞扫描
- 权限最小化原则
- 生产环境手动部署

## 📞 支持

- 项目维护者: dev-team@company.com
- DevOps 支持: devops@company.com
- GitLab 管理员: gitlab-admin@company.com
EOF

# 设置文件权限
chmod +x scripts/*.sh

echo -e "${GREEN}✅ GitLab CI/CD 配置文件创建完成${NC}"
echo ""
echo -e "${BLUE}📁 创建的文件:${NC}"
echo "  - .gitlab/ci-variables.yml          # CI/CD 变量模板"
echo "  - docker-compose.gitlab.yml         # Docker 开发环境"
echo "  - gitlab-runner-config.toml         # Runner 配置"
echo "  - scripts/gitlab-deploy.sh          # 部署脚本"
echo "  - scripts/setup-gitlab-variables.sh # 变量设置脚本"
echo "  - scripts/register-gitlab-runner.sh # Runner 注册脚本"
echo "  - README-GitLab.md                  # GitLab 项目文档"
echo ""
echo -e "${BLUE}🚀 下一步操作:${NC}"
echo "1. 查看 .gitlab/ci-variables.yml 并在 GitLab 项目中设置变量"
echo "2. 使用 scripts/register-gitlab-runner.sh 注册 Runner"
echo "3. 使用 scripts/setup-gitlab-variables.sh 设置 CI/CD 变量"
echo "4. 推送代码触发 CI/CD 流水线"
echo ""
echo -e "${GREEN}🎉 GitLab CI/CD 设置完成！${NC}"
