package gosshtun

import (
	"context"
	"net"
	"strconv"

	"github.com/armon/go-socks5"
	"github.com/fimreal/goutils/ezap"
)

func (st *SSHTun) Serve() {
	listenAddr := st.config.Tunnel.ListenAddr + ":" + strconv.Itoa(st.config.Tunnel.ListenPort)
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		ezap.Fatal(err)
	}

	go st.HandleSocks5(l)
	ezap.Info("Service listen on ", listenAddr)
	// for {
	// 	client, err := l.Accept()
	// 	if err != nil {
	// 		ezap.Error(err.Error())
	// 		continue
	// 	}
	// 	go handle(client)
	// }
}

// func tunnel(client net.Conn, server net.Conn) {
// 	go func() {
// 		io.Copy(client, server)
// 	}()
// 	io.Copy(server, client)
// }

// func handle(c net.Conn) {}

func (st *SSHTun) HandleSocks5(l net.Listener) {
	if st.client == nil {
		ezap.Errorf("未初始化 ssh 连接")
		return
	}

	conf := &socks5.Config{
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return st.client.Dial(network, addr)
		},
	}

	serverSocks, err := socks5.New(conf)
	if err != nil {
		ezap.Error(err.Error())
		return
	}
	if err := serverSocks.Serve(l); err != nil {
		ezap.Error("failed to create socks5 server", err)
	}
	// if err := serverSocks.ListenAndServe("tcp", listenAddr); err != nil {
	// 	ezap.Error("failed to create socks5 server", err)
	// }
}

func (st *SSHTun) Close() {
	if st.client != nil {
		st.client.Close()
	}
}
