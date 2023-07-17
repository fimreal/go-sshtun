package gosshtun

import (
	"bufio"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/fimreal/goutils/ezap"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// 隧道配置
type SSHConfig struct {
	// ssh server address, default: "127.0.0.1:22"
	RemoteAddr string
	// ssh user, default: "root"
	User string
	// ssh password, default: ""
	Password    string
	IdentityKey string
	// ssh config dir, default: {UserHomeDir}/.ssh
	IdentityKeyDir string
	// Timeout: 10s
	Timeout time.Duration
}

func DefaultSSHConfig() *SSHConfig {
	return &SSHConfig{
		RemoteAddr:     "",
		User:           "",
		Password:       "",
		IdentityKey:    "",
		IdentityKeyDir: "",
		Timeout:        10 * time.Second,
	}
}

func (c *SSHConfig) NewSSHClient() (*ssh.Client, error) {
	// ssh 服务连接地址
	host := c.SetSSHHost()
	user := c.SetSSHUser()
	// 获取验证方法
	var auth []ssh.AuthMethod
	if keys := c.ParseIdentityKey(); len(keys) != 0 {
		auth = append(auth, keys...)
	}
	if password := c.SSHPassword(); password != nil {
		auth = append(auth, password)
	}

	return ssh.Dial(
		"tcp",
		host,
		&ssh.ClientConfig{
			User:            user,
			Auth:            auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         c.Timeout,
		})
}

func (c *SSHConfig) SetSSHHost() string {
	host := c.RemoteAddr
	if host == "" {
		ezap.Printf("ssh server address : ")
		remote, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			ezap.Fatal(err.Error())
		}
		remote = strings.TrimSuffix(remote, "\n")
		remote = strings.TrimSuffix(remote, "\r")
		if !strings.Contains(remote, ":") {
			host = remote + ":22"
		} else {
			host = remote
		}
	}
	ezap.Debugf("set remote host: %s", host)
	return host
}

func (c *SSHConfig) SetSSHUser() string {
	user := c.User
	if user == "" {
		ezap.Printf("ssh user: ")
		username, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			ezap.Fatal(err.Error())
		}
		username = strings.TrimSuffix(username, "\n")
		username = strings.TrimSuffix(username, "\r")
		user = username
	}
	ezap.Debugf("set user: %s", user)
	return user
}

func (c *SSHConfig) SSHPassword() ssh.AuthMethod {
	stdin := int(syscall.Stdin)
	oldState, err := term.GetState(stdin)
	if err == nil {
		defer term.Restore(stdin, oldState)
	}
	if c.Password != "" {
		return ssh.Password(c.Password)
	} else {
		defer ezap.Println()
	}
	ezap.Printf("ssh password (press enter skip to using private Key): ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
	return ssh.Password(string(bytePassword))
}

// 从私钥文件创建 SSH 认证方法
func (c *SSHConfig) ParseIdentityKey() []ssh.AuthMethod {
	var auth []ssh.AuthMethod
	if key := parseIdentityFile(c.IdentityKey); key != nil {
		auth = append(auth, key)
	}
	if keys := parseIdentityDir(c.IdentityKeyDir); len(keys) != 0 {
		auth = append(auth, keys...)
	}
	return auth
}

func parseIdentityFile(keyFile string) ssh.AuthMethod {
	if keyFile == "" {
		return nil
	}
	byteKey, err := os.ReadFile(keyFile)
	if err != nil {
		ezap.Debugf("Failed to read private key file: %s", err)
		return nil
	}

	key, err := ssh.ParsePrivateKey(byteKey)
	if err != nil {
		ezap.Debugf("Failed to parse private key[%s]: %s", keyFile, err)
		return nil
	}

	return ssh.PublicKeys(key)
}

func parseIdentityDir(sshDir string) []ssh.AuthMethod {
	// 如果未配置目录，则在默认家目录下 .ssh 目录中查找
	if sshDir == "" {
		sshDir = homeSSHDir()
	}

	dirEntries, err := os.ReadDir(sshDir)
	if err != nil {
		ezap.Debugf("Could not get ssh config dir[%s]: %s", sshDir, err)
		return nil
	}

	var keys []ssh.AuthMethod
	for _, entry := range dirEntries {
		filename := entry.Name()
		if !entry.IsDir() && !strings.HasSuffix(filename, ".pub") && filename != "authorized_keys" && filename != "known_hosts" && filename != "config" {
			privateKey := path.Join(sshDir, filename)
			auth := parseIdentityFile(privateKey)
			if auth != nil {
				keys = append(keys, auth)
			}
		}
	}
	return keys
}

func homeSSHDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		ezap.Debugf("Could not get user home: %s, backoff use \".ssh\"", err)
		return ".ssh"
	}
	return path.Join(homeDir, ".ssh")
}
