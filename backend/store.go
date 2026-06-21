package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB
var dbMu sync.RWMutex

// openDB 打开或创建 SQLite 数据库，初始化表结构
func openDB() error {
	var err error
	db, err = sql.Open("sqlite", dbPath())
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS kv (k TEXT PRIMARY KEY, v TEXT)`); err != nil {
		return err
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS logs (id INTEGER PRIMARY KEY AUTOINCREMENT, time TEXT NOT NULL, msg TEXT NOT NULL)`); err != nil {
		return err
	}
	db.QueryRow("SELECT COUNT(*) FROM logs").Scan(&logCount)
	return nil
}

func closeDB() { db.Close() }

// ---- KV 表 CRUD ----

func kvGet(key string) string {
	dbMu.RLock()
	defer dbMu.RUnlock()
	var v string
	db.QueryRow("SELECT v FROM kv WHERE k=?", key).Scan(&v)
	return v
}

func kvSet(key, value string) {
	dbMu.Lock()
	defer dbMu.Unlock()
	db.Exec("INSERT OR REPLACE INTO kv(k,v) VALUES(?,?)", key, value)
}

// ---- logs 表 CRUD ----

var (
	logCount   int // 内存计数器，避免每行日志都 SELECT COUNT(*)
	logCountMu sync.Mutex
)

func logInsert(time, msg string) {
	dbMu.Lock()
	db.Exec("INSERT INTO logs(time,msg) VALUES(?,?)", time, msg)
	dbMu.Unlock()
}

// logQuery 查询最近 N 条日志（倒序）
func logQuery(limit int) []map[string]string {
	dbMu.RLock()
	defer dbMu.RUnlock()
	rows, err := db.Query("SELECT time,msg FROM logs ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var r []map[string]string
	for rows.Next() {
		var t, m string
		rows.Scan(&t, &m)
		r = append(r, map[string]string{"time": t, "msg": m})
	}
	return r
}

func logClear() {
	dbMu.Lock()
	db.Exec("DELETE FROM logs")
	dbMu.Unlock()
	logCountMu.Lock()
	logCount = 0
	logCountMu.Unlock()
}

// logTrim 超过上限时裁剪到一半
func logTrim(max int) {
	logCountMu.Lock()
	logCount++
	n := logCount
	logCountMu.Unlock()
	if n > max {
		logCountMu.Lock()
		logCount = max / 2
		logCountMu.Unlock()
		dbMu.Lock()
		db.Exec("DELETE FROM logs WHERE id NOT IN (SELECT id FROM logs ORDER BY id DESC LIMIT ?)", max/2)
		dbMu.Unlock()
	}
}

// ---- 运行时状态 ----

// runtimeState 服务运行时状态
type runtimeState struct {
	mu         sync.RWMutex
	startTime  time.Time
	lastUpload string
}

var state = &runtimeState{}

func (s *runtimeState) setStart() {
	s.startTime = time.Now()
}

func (s *runtimeState) setUpload(t string) {
	s.mu.Lock()
	s.lastUpload = t
	s.mu.Unlock()
}

func (s *runtimeState) snapshot() statusSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return statusSnapshot{
		Running:    true,
		StartTime:  s.startTime.Format("2006-01-02 15:04:05"),
		LastUpload: s.lastUpload,
	}
}

// restoreState 从 KV 表恢复 lastUpload
func restoreState() {
	state.lastUpload = kvGet("last_upload")
}

// saveUpload 更新上传时间（内存 + 持久化）
func saveUpload(t string) {
	state.setUpload(t)
	kvSet("last_upload", t)
}
