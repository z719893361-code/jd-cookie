package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const listenAddr = "127.0.0.1:17320"

var apiToken string

func initToken() string {
	if apiToken != "" {
		return apiToken
	}
	b := make([]byte, 16)
	rand.Read(b)
	apiToken = hex.EncodeToString(b)
	return apiToken
}

type apiResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func respOK(data ...interface{}) apiResp {
	if len(data) == 0 {
		return apiResp{Code: 0, Msg: "ok"}
	}
	return apiResp{Code: 0, Msg: "ok", Data: data[0]}
}

func respErr(msg string) apiResp {
	return apiResp{Code: 1, Msg: msg}
}

type statusSnapshot struct {
	Running    bool   `json:"running"`
	StartTime  string `json:"start_time"`
	LastUpload string `json:"last_upload"`
}

func writeResp(w http.ResponseWriter, r apiResp) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(r)
}

// authMux 全局 token 校验，无 token 直接 403
func authMux(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiToken == "" {
			next.ServeHTTP(w, r)
			return
		}
		t := r.Header.Get("X-Token")
		if t == "" {
			t = r.URL.Query().Get("token")
		}
		if t != apiToken {
			w.WriteHeader(403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Token")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func httpListen() error {
	initToken()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, r *http.Request) {
		writeResp(w, respOK(state.snapshot()))
	})

	mux.HandleFunc("GET /api/cookie", func(w http.ResponseWriter, r *http.Request) {
		cr := readCookie()
		if cr.OK {
			logf("读取Cookie  %s  %s", cr.Pin, cr.Key)
		} else {
			logf("读取Cookie失败 - %s", cr.Msg)
		}
		writeResp(w, respOK(cr))
	})

	mux.HandleFunc("POST /api/cookie", func(w http.ResponseWriter, r *http.Request) {
		var body struct{ Cookie string `json:"cookie"` }
		data, _ := io.ReadAll(r.Body)
		cfg := loadConfig()
		var qr qlResult
		if len(data) > 0 && json.Unmarshal(data, &body) == nil && body.Cookie != "" {
			qr = uploadCookie(cfg, body.Cookie)
		} else {
			qr = uploadFromDB(cfg)
		}
		if qr.OK {
			saveUpload(time.Now().Format("2006-01-02 15:04:05"))
			logf("[手动] 上传成功")
		} else {
			logf("[手动] 上传失败 - %s", qr.Msg)
		}
		writeResp(w, respOK(qr))
	})

	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		writeResp(w, respOK(loadConfig()))
	})
	mux.HandleFunc("PUT /api/config", func(w http.ResponseWriter, r *http.Request) {
		var cfg Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeResp(w, respErr("JSON 解析失败"))
			return
		}
		if cfg.EnvName == "" {
			writeResp(w, respErr("请填写环境变量名称"))
			return
		}
		if !cfg.isValid() {
			writeResp(w, respErr("请填写面板地址、用户名和密码"))
			return
		}
		saveConfig(&cfg)
		logf("保存配置  %s  %s:***@%s", cfg.EnvName, cfg.QLUser, cfg.QLURL)
		writeResp(w, respOK())
	})

	mux.HandleFunc("POST /api/test", func(w http.ResponseWriter, r *http.Request) {
		var tmp Config
		cfg := loadConfig()
		if r.Body != nil {
			body, _ := io.ReadAll(r.Body)
			if len(body) > 0 && json.Unmarshal(body, &tmp) == nil {
				if tmp.QLURL != "" {
					cfg.QLURL = tmp.QLURL
				}
				if tmp.QLUser != "" {
					cfg.QLUser = tmp.QLUser
				}
				if tmp.QLPass != "" {
					cfg.QLPass = tmp.QLPass
				}
				if tmp.EnvName != "" {
					cfg.EnvName = tmp.EnvName
				}
			}
		}
		qr := testQL(cfg)
		if qr.OK {
			logf("测试连接成功  %s", cfg.QLURL)
			writeResp(w, respOK())
		} else {
			logf("测试连接失败 - %s", qr.Msg)
			writeResp(w, respErr(qr.Msg))
		}
	})

	mux.HandleFunc("GET /api/log/stream", func(w http.ResponseWriter, r *http.Request) {
		fl, ok := w.(http.Flusher)
		if !ok {
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		rows := logQuery(1000)
		for i := len(rows) - 1; i >= 0; i-- {
			b, _ := json.Marshal(rows[i])
			fmt.Fprintf(w, "data: %s\n\n", b)
			fl.Flush()
		}

		ch := subscribe()
		defer unsubscribe(ch)
		for {
			select {
			case <-r.Context().Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", msg)
				fl.Flush()
			}
		}
	})

	mux.HandleFunc("DELETE /api/log", func(w http.ResponseWriter, r *http.Request) {
		logClear()
		writeResp(w, respOK())
	})

	return http.ListenAndServe(listenAddr, cors(authMux(mux)))
}
