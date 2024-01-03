package main

import (
	"errors"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	pflag.BoolP("reset", "R", false, "reset/prune system proxy rule.")
	// service
	pflag.BoolP("debug", "d", false, "debug mode")
	pflag.BoolP("sysproxy", "S", false, "enable system proxy")
	pflag.StringP("pac", "P", "", "enable pac (proxy auto-config). need a pac url, like \"http://127.0.0.1:8000/my.pac\", or use embedded rules(gfw, tiny)")
	pflag.StringP("listen", "l", "127.0.0.1:1080", "service listing on")
	pflag.StringP("auth", "a", "", "enable and set socks5/http proxy service authentication, eg. \"user:pass\"")

	// ssh
	pflag.StringP("host", "h", "", "ssh server address, like \"192.168.50.220:2222\"")
	pflag.StringP("username", "u", "", "ssh user")
	pflag.StringP("password", "p", "", "ssh password")
	pflag.StringP("identitykey", "i", "", "identity key file")
	pflag.StringP("identitykeydir", "k", "", "identity key dir")
	pflag.DurationP("timeout", "t", 10*time.Second, "timeout")

	// viper parse
	viper.AutomaticEnv()
	pflag.ErrHelp = errors.New("")
	pflag.Parse()
	// pflag sort is no used
	// pflag.CommandLine.SortFlags = false
	viper.BindPFlags(pflag.CommandLine)
}
