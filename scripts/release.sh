#!/bin/bash

# 发布脚本 - 自动创建版本标签并触发 CI/CD

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 获取当前版本
CURRENT_VERSION=$(cat VERSION)
echo -e "${YELLOW}Current version: ${CURRENT_VERSION}${NC}"

# 版本类型参数
VERSION_TYPE=${1:-patch}

# 解析版本号
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

# 根据类型增加版本号
case $VERSION_TYPE in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo -e "${RED}Invalid version type. Use: major, minor, or patch${NC}"
        exit 1
        ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo -e "${GREEN}New version: ${NEW_VERSION}${NC}"

# 确认发布
read -p "Do you want to release version ${NEW_VERSION}? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Release cancelled${NC}"
    exit 0
fi

# 检查工作目录是否干净
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${RED}Working directory is not clean. Please commit or stash changes.${NC}"
    exit 1
fi

# 更新版本文件
echo "$NEW_VERSION" > VERSION

# 更新 README 中的版本信息（如果存在）
if [[ -f README.md ]]; then
    sed -i.bak "s/version-[0-9]\+\.[0-9]\+\.[0-9]\+/version-${NEW_VERSION}/g" README.md
    rm -f README.md.bak
fi

# 提交版本更新
git add VERSION README.md
git commit -m "chore: bump version to ${NEW_VERSION}"

# 创建标签
git tag -a "v${NEW_VERSION}" -m "Release version ${NEW_VERSION}"

# 推送到远程仓库
echo -e "${YELLOW}Pushing to remote repository...${NC}"
git push origin main
git push origin "v${NEW_VERSION}"

echo -e "${GREEN}✅ Release ${NEW_VERSION} created successfully!${NC}"
echo -e "${YELLOW}GitLab CI will now build and push the Docker images.${NC}"
echo -e "${YELLOW}Check the pipeline at: ${CI_PIPELINE_URL:-https://gitlab.com/your-project/-/pipelines}${NC}"