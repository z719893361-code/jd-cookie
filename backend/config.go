package main

import "net/url"

// Config 青龙面板连接配置
type Config struct {
	QLURL   string `json:"ql_url"`
	QLUser  string `json:"ql_user"`
	QLPass  string `json:"ql_pass"`
	EnvName string `json:"env_name"`
}

func (c *Config) isValid() bool {
	if c.QLURL == "" || c.QLUser == "" || c.QLPass == "" {
		return false
	}
	u, err := url.Parse(c.QLURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}
	return true
}

// loadConfig 从 KV 表读取青龙面板配置
func loadConfig() *Config {
	c := &Config{
		QLURL:   kvGet("ql_url"),
		QLUser:  kvGet("ql_user"),
		QLPass:  kvGet("ql_pass"),
		EnvName: kvGet("env_name"),
	}
	if c.EnvName == "" {
		c.EnvName = "JD_COOKIE"
	}
	return c
}

// saveConfig 将青龙面板配置写入 KV 表
func saveConfig(c *Config) {
	kvSet("ql_url", c.QLURL)
	kvSet("ql_user", c.QLUser)
	kvSet("ql_pass", c.QLPass)
	kvSet("env_name", c.EnvName)
}
