#!/system/bin/sh
MODDIR="${0%/*}"
export TZ=Asia/Shanghai

TOKEN=$(head -c16 /dev/urandom | xxd -p)
echo "$TOKEN" > "$MODDIR/webroot/token.txt"

exec "$MODDIR/bin/jd-cookie" daemon \
    --token "$TOKEN" \
    --db "$MODDIR/data.db" \
    --cookie-db "/data/data/com.jingdong.app.mall/app_webview/Default/Cookies"
