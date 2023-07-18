package main

import (
	"errors"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	// service
	pflag.BoolP("debug", "d", false, "debug mode")
	pflag.StringP("listen", "l", "127.0.0.1:10080", "service listing on")
	pflag.BoolP("sysproxy", "s", false, "enable system proxy")
	pflag.StringP("pac", "P", "", "pac url, like \"http://127.0.0.1:8000/my.pac\"")

	// ssh
	pflag.StringP("host", "h", "", "ssh server address, like \"epurs.com:2222\"")
	pflag.StringP("user", "u", "", "ssh user")
	pflag.StringP("password", "p", "", "ssh password")
	pflag.StringP("identitykey", "i", "", "identity key file")
	pflag.StringP("identitykeydir", "k", "", "identity key dir")
	pflag.DurationP("timeout", "t", 10*time.Second, "timeout")

	viper.AutomaticEnv()
	pflag.ErrHelp = errors.New("")
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}
