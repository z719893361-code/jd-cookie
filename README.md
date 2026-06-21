# jd-cookie

KernelSU 模块 - 自动读取京东 Cookie 并同步至青龙面板。

## 功能

- 定时从京东 App WebView Cookie 数据库读取 `pt_key` / `pt_pin`
- Cookie 去重后自动上传至青龙面板环境变量
- Go 守护进程，内存占用 < 10MB
- WebUI 配置管理 + 实时 SSE 日志流
- Token 鉴权，外部请求一律拒绝

## 安装

1. 下载 [releases 页面](https://github.com/z719893361-code/jd-cookie/releases) 的 `jd_assistant.zip`
2. 在 KernelSU Manager 中刷入模块
3. 重启，WebUI 自动加载

## 使用

1. 打开京东 App 登录账号（让 Cookie 写入 WebView 数据库）
2. 打开 KernelSU Manager → 模块 → 京东助手 → 打开
3. 填写青龙面板地址、用户名、密码，保存
4. 点击"读取"获取 Cookie，或等待每 10 分钟自动上传

## 结构

```
├── backend/        Go 守护进程
│   ├── config.go   青龙配置
│   ├── cookie.go   Cookie 读取
│   ├── log.go      日志系统
│   ├── main.go     入口 + 守护循环
│   ├── paths.go    路径解析
│   ├── ql.go       青龙 API
│   ├── server.go   HTTP 路由
│   └── store.go    SQLite 存储
├── frontend/       Vue 3 WebUI
├── kernelsu/       模块打包
│   ├── service.sh  启动脚本
│   └── module.prop
└── .gitignore
```

## 构建

```bash
# 后端
cd backend && CGO_ENABLED=0 GOOS=android GOARCH=arm64 go build -ldflags="-s -w" -o ../kernelsu/bin/jd-cookie .

# 前端
cd frontend && npm install && npm run build
```

## License

MIT
