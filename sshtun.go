package gosshtun

import (
	"golang.org/x/crypto/ssh"
)

type SSHTun struct {
	Client     *ssh.Client
	ListenAddr string
}

func NewSSHTun(c *SSHConfig) (*SSHTun, error) {
	sc, err := c.NewSSHClient()
	return &SSHTun{
		Client:     sc,
		ListenAddr: "0.0.0.0:10080",
	}, err
}

func (st *SSHTun) Close() {
	if st.Client != nil {
		st.Client.Close()
	}
}
