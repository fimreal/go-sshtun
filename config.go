package gosshtun

import "time"

// 配置
type Config struct {
	SSHConfig
	Tunnel
}

// ssh 连接配置
type SSHConfig struct {
	RemoteAddr     string
	RemotePort     int
	Username       string
	Password       string
	IdentityKey    string
	IdentityKeyDir string
	Timeout        time.Duration
}

// 动态隧道监听地址配置
type Tunnel struct {
	ListenAddr string
	ListenPort int
}

func NewConfig() *Config {
	return &Config{
		SSHConfig: SSHConfig{
			RemoteAddr:     "127.0.0.1",
			RemotePort:     22,
			Username:       "root",
			Password:       "",
			IdentityKey:    "",
			IdentityKeyDir: "",
			Timeout:        10 * time.Second,
		},
		Tunnel: Tunnel{
			ListenAddr: "0.0.0.0",
			ListenPort: 10080,
		},
	}
}
