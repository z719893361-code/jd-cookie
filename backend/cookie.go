package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
)

// cookieData 京东 Cookie 读取结果
type cookieData struct {
	OK     bool   `json:"ok"`
	Msg    string `json:"msg"`
	Key    string `json:"key,omitempty"`
	Pin    string `json:"pin,omitempty"`
	Cookie string `json:"cookie,omitempty"`
}

// readCookie 从京东 WebView 的 Cookie 数据库中读取 pt_key 和 pt_pin
func readCookie() cookieData {
	path := cookieDBPath()
	if _, err := os.Stat(path); err != nil {
		return cookieData{OK: false, Msg: "数据库不存在，请先打开京东"}
	}
	d, err := sql.Open("sqlite", "file:"+path+"?mode=ro&_pragma=busy_timeout=3000")
	if err != nil {
		return cookieData{OK: false, Msg: "打开失败"}
	}
	defer d.Close()

	var key, pin string
	err = d.QueryRow("SELECT value FROM cookies WHERE host_key='.jd.com' AND name='pt_key' LIMIT 1").Scan(&key)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return cookieData{OK: false, Msg: "查询 pt_key 失败"}
	}
	err = d.QueryRow("SELECT value FROM cookies WHERE host_key='.jd.com' AND name='pt_pin' LIMIT 1").Scan(&pin)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return cookieData{OK: false, Msg: "查询 pt_pin 失败"}
	}
	if key == "" || pin == "" {
		return cookieData{OK: false, Msg: "Cookie 为空"}
	}
	return cookieData{OK: true, Msg: "读取成功", Key: key, Pin: pin, Cookie: fmt.Sprintf("pt_key=%s;pt_pin=%s;", key, pin)}
}
