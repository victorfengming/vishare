#!/bin/bash
# 在 Mac (Apple Silicon) 上运行此脚本来构建 vishare 服务端
set -e

echo "==> 检查依赖..."
if ! command -v go &>/dev/null; then
    echo "错误: 未找到 Go，请先安装 https://go.dev/dl/"
    exit 1
fi

echo "==> Go 版本: $(go version)"

echo "==> 安装系统依赖（需要 Homebrew）..."
if command -v brew &>/dev/null; then
    # robotgo / gohook 在 macOS 上不需要额外的系统库，Xcode CLT 即可
    xcode-select --install 2>/dev/null || true
else
    echo "警告: 未找到 Homebrew，如果编译失败请先安装 Xcode Command Line Tools:"
    echo "  xcode-select --install"
fi

echo "==> 下载依赖..."
GOPROXY=direct go mod download

echo "==> 编译 macOS ARM64 服务端..."
mkdir -p build
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
    go build -o build/vishare-darwin-arm64 ./cmd/vishare

echo ""
echo "✅ 编译完成: build/vishare-darwin-arm64"
echo ""
echo "运行方式（macOS 需要辅助功能权限）:"
echo "  ./build/vishare-darwin-arm64 --config config.server.toml"
echo ""
echo "首次运行时，macOS 会提示授予「辅助功能」和「输入监控」权限，"
echo "在「系统设置 → 隐私与安全性」中允许即可。"
