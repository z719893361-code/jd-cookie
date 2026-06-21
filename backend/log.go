package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const maxLogRows = 10000

// SSE 订阅者管理
var (
	subscribers = make(map[chan string]struct{})
	subMu       sync.Mutex
)

// subscribe 注册 SSE 订阅通道
func subscribe() chan string {
	ch := make(chan string, 128)
	subMu.Lock()
	subscribers[ch] = struct{}{}
	subMu.Unlock()
	return ch
}

// unsubscribe 移除 SSE 订阅通道
func unsubscribe(ch chan string) {
	subMu.Lock()
	delete(subscribers, ch)
	subMu.Unlock()
}

// logf 写日志：写入 SQLite 同时广播到 SSE 订阅者
func logf(format string, args ...interface{}) {
	now := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	logInsert(now, msg)

	line, _ := json.Marshal(map[string]string{"time": now, "msg": msg})
	subMu.Lock()
	for ch := range subscribers {
		select {
		case ch <- string(line):
		default:
		}
	}
	subMu.Unlock()

	logTrim(maxLogRows)
}
