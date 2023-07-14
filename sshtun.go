package gosshtun

import (
	"golang.org/x/crypto/ssh"
)

type SSHTun struct {
	client *ssh.Client
	config *Config
}

func NewSSHTun(config *Config) (*SSHTun, error) {
	sc, err := NewSSHClient(config)
	return &SSHTun{
		client: sc,
		config: config,
	}, err
}
