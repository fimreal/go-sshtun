package gosshtun

import (
	"golang.org/x/crypto/ssh"
)

type SSHTun struct {
	Client *ssh.Client
	Config *SSHConfig
	// 动态隧道监听地址配置
	ListenAddr string
}

func NewSSHTun(c *SSHConfig) (*SSHTun, error) {
	sc, err := c.NewSSHClient()
	return &SSHTun{
		Client:     sc,
		Config:     c,
		ListenAddr: "0.0.0.0:10080",
	}, err
}
