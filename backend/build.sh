#!/bin/bash
# ============================================================
# jd-cookie 编译脚本
# 在 WSL 中执行: bash build.sh
# 交叉编译 ARM64 Android 静态二进制
# 输出: ../kernelsu/bin/jd-cookie
# ============================================================
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT="$SCRIPT_DIR/../kernelsu/bin/jd-cookie"

echo "源码目录: $SCRIPT_DIR"
echo "输出: $OUT"

cd "$SCRIPT_DIR"

# 下载依赖
echo "--- 下载依赖 ---"
go mod tidy 2>&1 | tail -5

# 交叉编译
echo "--- 编译 (GOOS=android GOARCH=arm64 CGO_ENABLED=0) ---"
GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o "$OUT" .

echo "--- 验证 ---"
ls -la "$OUT"
file "$OUT"

# 尝试 upx 压缩（可选）
if command -v upx >/dev/null 2>&1; then
    echo "--- upx 压缩 ---"
    upx --best "$OUT"
    ls -la "$OUT"
else
    echo "(upx 未安装，跳过压缩)"
fi

echo "完成！"
