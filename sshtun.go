package gosshtun

import (
	"strconv"

	"github.com/fimreal/goutils/ezap"
	"golang.org/x/crypto/ssh"
)

type SSHTun struct {
	Client        *ssh.Client
	ListenAddr    string
	TotalUpload   int64
	TotalDownload int64
}

func NewSSHTun(c *SSHConfig) (*SSHTun, error) {
	sc, err := c.NewSSHClient()
	return &SSHTun{
		Client: sc,
		// ListenAddr: "0.0.0.0:1080",
		TotalUpload:   0,
		TotalDownload: 0,
	}, err
}

func (st *SSHTun) Close() {
	if st.Client != nil {
		st.Client.Close()
	}
}

func (st *SSHTun) Stat() {
	ezap.Println("Statistic:")
	ezap.Info("Total Upload: ", beautifySize(st.TotalUpload))
	ezap.Info("Total Download: ", beautifySize(st.TotalDownload))
}

// https://github.com/justmao945/mallory/blob/ad32fd8abd0c4a763734717e762e2afae187fce5/beautify.go#L24
func beautifySize(s int64) string {
	switch {
	case s < 1024:
		return strconv.FormatInt(s, 10) + "B"
	case s < 1024*1024:
		return strconv.FormatInt(s/1024, 10) + "KB"
	case s < 1024*1024*1024:
		return strconv.FormatInt(s/1024/1024, 10) + "MB"
	default:
		return strconv.FormatInt(s/1024/1024/1024, 10) + "GB"
	}
}
