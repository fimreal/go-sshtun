package gosshtun

import (
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/fimreal/goutils/ezap"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func NewSSHClient(config *Config) (*ssh.Client, error) {
	// ssh 服务连接地址
	remote := config.SSHConfig.RemoteAddr + ":" + strconv.Itoa(config.SSHConfig.RemotePort)
	// 获取验证方法
	var auth []ssh.AuthMethod
	if key := ParseIdentityFile(config.SSHConfig.IdentityKey); key != nil {
		auth = append(auth, key)
	}
	if keys := ParseIdentityDir(config.SSHConfig.IdentityKeyDir); len(keys) != 0 {
		auth = append(auth, keys...)
	}
	if pass := SSHPassword(config.SSHConfig.Password); pass != nil {
		auth = append(auth, pass)
	}
	return ssh.Dial(
		"tcp",
		remote,
		&ssh.ClientConfig{
			User:            config.SSHConfig.Username,
			Auth:            auth,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         config.SSHConfig.Timeout,
		})
}

func SSHPassword(password string) ssh.AuthMethod {
	if password != "" {
		return ssh.Password(password)
	}
	ezap.Printf("ssh password (press enter skip): ")
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
	ezap.Println()
	return ssh.Password(string(bytePassword))
}

// 从私钥文件创建 SSH 认证方法
func ParseIdentityFile(file string) ssh.AuthMethod {
	if file == "" {
		return nil
	}
	byteKey, err := os.ReadFile(file)
	if err != nil {
		ezap.Warn("Failed to read private key file: %s", err)
		return nil
	}

	key, err := ssh.ParsePrivateKey(byteKey)
	if err != nil {
		ezap.Errorf("Failed to parse private key[%s]: %s", file, err)
		return nil
	}

	return ssh.PublicKeys(key)
}

// 使用目录中存在的私钥
func ParseIdentityDir(sshDir string) (keys []ssh.AuthMethod) {
	// 如果未配置目录，则在默认家目录下 .ssh 目录中查找
	if sshDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			ezap.Warn("Could not get user home: %s", err)
			return nil
		}
		sshDir = path.Join(homeDir, ".ssh")
	}

	dirEntries, err := os.ReadDir(sshDir)
	if err != nil {
		ezap.Warn("Could not get ssh config dir: %s", err)
		return nil
	}

	for _, entry := range dirEntries {
		filename := entry.Name()
		if !entry.IsDir() && !strings.HasSuffix(filename, ".pub") && filename != "authorized_keys" && filename != "known_hosts" && filename != "config" {
			privateKey := path.Join(sshDir, filename)
			auth := ParseIdentityFile(privateKey)
			if auth != nil {
				keys = append(keys, auth)
			}
		}
	}
	return
}
