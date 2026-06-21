package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/spf13/cobra"
)

func init() {
	// 强制东八区（Android 默认 UTC，TZ 环境变量可能不生效）
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		time.Local = loc
	}
}

// pidAlive 检查进程是否存活（仅 Linux）
func pidAlive(pid int) bool {
	_, err := os.Stat("/proc/" + strconv.Itoa(pid))
	return pid > 0 && err == nil
}

// lastCookie 上次成功上传的 Cookie 值，用于去重（内存常驻，重启从 SQLite 恢复）
var lastCookie string

// runDaemon 服务主循环：启动 HTTP 服务，每 10 分钟自动读取并上传 Cookie
func runDaemon() {
	kvSet("pid", strconv.Itoa(os.Getpid()))
	defer kvSet("pid", "")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sigCh; kvSet("pid", ""); os.Exit(0) }()

	logf("服务已启动  PID %d", os.Getpid())

	go func() {
		logf("HTTP 监听  %s", listenAddr)
		if err := httpListen(); err != nil {
			logf("HTTP 异常 - %v", err)
		}
	}()

	doCycle := func() {
		defer func() {
			if r := recover(); r != nil {
				logf("循环异常 - %v", r)
			}
		}()

		cfg := loadConfig()
		if !cfg.isValid() {
			logf("跳过 - 青龙配置不完整")
			return
		}

		logf("[自动] 开始执行")
		t0 := time.Now()
		cr := readCookie()
		if !cr.OK {
			logf("读取失败 - %s", cr.Msg)
		} else {
			logf("读取成功  %s  %s  %.0fms", cr.Pin, cr.Key, sinceMs(t0))

			if cr.Cookie == lastCookie {
				logf("Cookie 未变化，跳过上传")
				return
			}

			qr := uploadCookie(cfg, cr.Cookie)
			if qr.OK {
				logf("[自动] 上传成功  %.0fms", sinceMs(t0))
				saveUpload(time.Now().Format("2006-01-02 15:04:05"))
				lastCookie = cr.Cookie
				kvSet("last_cookie", cr.Cookie)
			} else {
				logf("[自动] 上传失败 - %s", qr.Msg)
			}
		}
	}

	// 启动后立即执行一次，之后每 10 分钟
	doCycle()
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		doCycle()
	}
}

// sinceMs 计算距 t 的毫秒数，用于日志性能计时
func sinceMs(t time.Time) float64 {
	return float64(time.Since(t).Microseconds()) / 1000
}

func main() {
	var dbFlag, cookieDBFlag, tokenFlag string

	root := &cobra.Command{
		Use:   "jd-cookie",
		Short: "京东助手 - 自动读取京东Cookie并同步至青龙面板",
	}

	daemon := &cobra.Command{
		Use:   "daemon",
		Short: "启动后台服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 设置访问令牌
			if tokenFlag != "" {
				apiToken = tokenFlag
			}
			// 覆盖默认路径
			if dbFlag != "" {
				dbPathOverride = dbFlag
				modDir = filepath.Dir(dbFlag)
			}
			if cookieDBFlag != "" {
				cookieDBOverride = cookieDBFlag
			}

			os.MkdirAll(modDir, 0755)
			if err := openDB(); err != nil {
				return fmt.Errorf("存储初始化失败: %w", err)
			}
			defer closeDB()

			// 从 SQLite 恢复上次状态
			restoreState()
			lastCookie = kvGet("last_cookie")

			// 单实例检查
			if pidStr := kvGet("pid"); pidStr != "" {
				if pid, err := strconv.Atoi(pidStr); err == nil && pidAlive(pid) && pid != os.Getpid() {
					return fmt.Errorf("已有进程运行中")
				}
			}
			runDaemon()
			return nil
		},
	}
	daemon.Flags().StringVar(&dbFlag, "db", "", "SQLite 数据库路径")
	daemon.Flags().StringVar(&cookieDBFlag, "cookie-db", "", "京东 Cookie 数据库路径")
	daemon.Flags().StringVar(&tokenFlag, "token", "", "API 访问令牌")

	version := &cobra.Command{
		Use:   "version",
		Short: "显示版本",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("jd-cookie v4.0.0")
		},
	}

	root.AddCommand(daemon, version)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
