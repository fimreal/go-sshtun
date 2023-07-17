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
	if viper.GetBool("debug") {
		ezap.SetLevel("debug")
	}

	// config := gosshtun.DefaultSSHConfig()
	config := &gosshtun.SSHConfig{
		RemoteAddr:     viper.GetString("host"),
		User:           viper.GetString("user"),
		Password:       viper.GetString("password"),
		IdentityKey:    viper.GetString("identitykey"),
		IdentityKeyDir: viper.GetString("identitykeydir"),
		Timeout:        viper.GetDuration("timeout"),
	}

	st, err := gosshtun.NewSSHTun(config)
	if err != nil {
		ezap.SetLogTime("")
		ezap.Fatal(err)
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
