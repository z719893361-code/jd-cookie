#!/system/bin/sh
# ============================================================
# 京东 Cookie 读取器 - customize.sh
# KernelSU 安装时执行：权限、SELinux
# ============================================================

SKIPUNZIP=0
AUTOMOUNT=true
print() { echo "- $1"; }

# 1. 目录权限
print "设置目录权限..."
chmod 755 "$MODPATH"
chmod 755 "$MODPATH/bin" 2>/dev/null
chmod 755 "$MODPATH/webroot" 2>/dev/null

# 2. Go 二进制权限和 SELinux
BIN="$MODPATH/bin/jd-cookie"
if [ -f "$BIN" ]; then
    chmod 755 "$BIN"
    chcon u:object_r:system_file:s0 "$BIN" 2>/dev/null || true
    print "jd-cookie 已就绪"
else
    print "错误: 未找到 bin/jd-cookie！模块可能损坏，请重新下载"
fi

# 3. service.sh 权限
[ -f "$MODPATH/service.sh" ] && chmod 755 "$MODPATH/service.sh"

print "安装完成！打开 WebUI 配置青龙面板，重启后服务自动运行"
