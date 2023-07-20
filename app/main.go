package main

import (
	"bufio"
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
	ezap.SetLogTime("2006-1-2 15:04:05")
	ezap.Debug("debug mod")

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

	// ssh tunnel service
	go st.Serve()

	// pac
	var pacon bool
	if pac := viper.GetString("pac"); pac != "" {
		pacon = st.PacOn(pac)
		st.PacInspect()
	}

	// system proxy
	var enabledSystemProxy bool
	if viper.GetBool("sysproxy") {
		enabledSystemProxy = st.EnableSystemProxy()
	}

	go func() {
		for {
			_, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				return
			}
			st.Stat()
			// ezap.Printf("\rTips: press CTRL + c to exit application")
		}
	}()

	// catch signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signalChan
	ezap.Println()
	// unset system proxy
	if enabledSystemProxy {
		st.DisableSystemProxy()
	}
	// pac off
	if pacon {
		st.PacOff()
	}
	// close ssh client
	st.Close()
	os.Exit(0)
}
