#!/system/bin/sh
MODDIR="${0%/*}"

# 读取手机时区
TZ=$(getprop persist.sys.timezone 2>/dev/null)
[ -z "$TZ" ] && TZ="Asia/Shanghai"

TOKEN=$(head -c16 /dev/urandom | xxd -p)
echo "$TOKEN" > "$MODDIR/webroot/token.txt"

exec "$MODDIR/bin/jd-cookie" daemon \
    --tz "$TZ" \
    --token "$TOKEN" \
    --db "$MODDIR/data.db" \
    --cookie-db "/data/data/com.jingdong.app.mall/app_webview/Default/Cookies"
