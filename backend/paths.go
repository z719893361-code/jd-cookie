package main

import (
	"os"
	"path/filepath"
)

// 模块根目录，默认从二进制路径推导（bin/jd-cookie → 模块根）
var modDir string

// 由启动参数覆盖，用于指定非默认位置
var dbPathOverride string
var cookieDBOverride string

func init() {
	exe, _ := os.Executable()
	modDir = filepath.Dir(filepath.Dir(exe))
}

// 本程序 SQLite 数据库路径
func dbPath() string {
	if dbPathOverride != "" {
		return dbPathOverride
	}
	return modDir + "/data.db"
}

// 京东 WebView Cookie 数据库路径
func cookieDBPath() string {
	if cookieDBOverride != "" {
		return cookieDBOverride
	}
	return "/data/data/com.jingdong.app.mall/app_webview/Default/Cookies"
}
