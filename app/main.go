package main

import (
	"os"
	"os/signal"
	"syscall"

	gosshtun "github.com/fimreal/go-sshtun"
	"github.com/fimreal/goutils/ezap"
	"github.com/spf13/viper"
)

func main() {
	// ezap.SetLevel("debug")

	config := gosshtun.DefaultSSHConfig()
	config.RemoteAddr = viper.GetString("host")
	config.User = viper.GetString("user")
	config.Password = viper.GetString("password")
	config.IdentityKey = viper.GetString("identitykey")
	config.IdentityKeyDir = viper.GetString("identitykeydir")
	config.Timeout = viper.GetDuration("timeout")

	st, err := gosshtun.NewSSHTun(config)
	if err != nil {
		ezap.SetLogTime("")
		ezap.Fatal(err)
		ezap.SetLogTime("2006-1-2 15:04:05")
	}
	st.ListenAddr = viper.GetString("listen")

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-signalChan
		st.Close()
		os.Exit(0)
	}()

	st.Serve()
}
