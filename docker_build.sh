#!/bin/bash
# 更健壮的错误处理（包括 pipefail）
set -Eeuo pipefail

# 确保日志目录存在
mkdir -p ./logs/docker

# 日志记录
log_file="./logs/docker/build_$(date +%Y%m%d_%H%M%S).log"
exec > >(tee -a "$log_file") 2>&1

echo "=== 开始构建流程 $(date) ==="

# ---- 可配置参数（也可以通过环境变量覆盖）----
BINARY_NAME="${BINARY_NAME:-downstream-server-manager}"  # 你的最终二进制名
TARGET_OS="${TARGET_OS:-linux}"                           # 目标 OS：linux
TARGET_ARCH="${TARGET_ARCH:-amd64}"                       # 目标架构：amd64 (可改为 arm64)
OUTPUT_DIR="${OUTPUT_DIR:-./bin}"                         # 输出目录
OUTPUT_PATH="${OUTPUT_DIR}/${BINARY_NAME}"                # 输出路径
IMAGE_TAG="${IMAGE_TAG:-downstream-server-manager:alpine}" # 镜像标签
# 版本信息（可注入到代码中）
GIT_REV=$(git rev-parse --short HEAD 2>/dev/null || echo "nogit")
BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
LD_FLAGS="${LD_FLAGS:--s -w -X main.gitRev=${GIT_REV} -X main.buildTime=${BUILD_TIME}}"

# Go 环境（启用 Modules + 允许自动拉取工具链）
export GO111MODULE=on
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"

# 是否静态编译
export CGO_ENABLED="${CGO_ENABLED:-0}"

echo "（可选）宿主机预下载 Go 模块..."
go mod download || true  # 即使失败也不终止

# 简单检查 Go 版本（你的 go.mod 若要求 >=1.24，建议本机 Go 版本 >=1.24 或设置 GOTOOLCHAIN=auto）
echo "Go 版本: $(go version || echo '未检测到 go')"
echo "GOTOOLCHAIN=${GOTOOLCHAIN}, CGO_ENABLED=${CGO_ENABLED}"

echo "开始交叉编译：GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} -> ${OUTPUT_PATH}"
mkdir -p "${OUTPUT_DIR}"

# 注意：CGO_ENABLED=0 时，Go 会生成适合 Alpine 运行的静态二进制（无需额外 libc）
GOOS="${TARGET_OS}" GOARCH="${TARGET_ARCH}" \
  go build -trimpath -buildvcs=false -ldflags="${LD_FLAGS}" \
  -o "${OUTPUT_PATH}" .

# 赋予执行权限（通常 go build 已设置，可再次确保）
chmod +x "${OUTPUT_PATH}"

# 打印二进制信息（可选）
echo "构建完成的二进制：${OUTPUT_PATH}"
command -v file >/dev/null 2>&1 && file "${OUTPUT_PATH}" || true

echo "开始构建 Docker 运行时镜像（使用 alpine 作为运行时镜像）..."
# 注意：dockerfile 是运行时专用（不含 builder 阶段）
docker build -f dockerfile -t "${IMAGE_TAG}" .

echo "=== 构建完成 $(date) ==="
echo "生成的镜像:"
docker images | grep -E "REPOSITORY|${IMAGE_TAG%:*}" || true

echo "提示: 可使用 'docker system prune' 清理未使用的镜像/缓存"
