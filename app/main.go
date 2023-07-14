package main

import (
	"os"
	"os/signal"
	"syscall"

	gosshtun "github.com/fimreal/go-sshtun"
	"github.com/fimreal/goutils/ezap"
)

func main() {
	config := gosshtun.NewConfig()
	config.RemoteAddr = "home.epurs.com"
	config.RemotePort = 22
	config.Username = "root"
	// config.Password = ""

	st, err := gosshtun.NewSSHTun(config)
	if err != nil {
		ezap.Fatal(err)
	}

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-signalChan
		st.Close()
		os.Exit(0)
	}()

	st.Serve()
}
