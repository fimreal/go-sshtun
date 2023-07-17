package gosshtun

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/fimreal/goutils/ezap"
)

func (st *SSHTun) Serve() {
	listenAddr := st.ListenAddr
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		ezap.Fatal(err)
	}

	ezap.Info("Service listen on ", listenAddr)
	for {
		client, err := l.Accept()
		if err != nil {
			ezap.Error(err.Error())
			continue
		}
		go st.handle(client)
	}
}

func (st *SSHTun) handle(client net.Conn) {
	var b [1024]byte
	_, err := client.Read(b[:])
	if err != nil {
		ezap.Error(err.Error())
		return
	}
	if b[0] == 0x05 { //only for socks5
		//response to client: no need to validation
		client.Write([]byte{0x05, 0x00})
		n, err := client.Read(b[:])
		if err != nil {
			ezap.Error(err.Error())
			return
		}
		var host, port string
		switch b[3] {
		case 0x01: //IP V4
			host = net.IPv4(b[4], b[5], b[6], b[7]).String()
		case 0x03: //domain name
			host = string(b[5 : n-2]) //b[4] length of domain name
		case 0x04: //IP V6
			host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
		}
		port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
		// server, err := net.Dial("tcp", net.JoinHostPort(host, port))
		server, err := st.Client.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			ezap.Error("[socks5]" + err.Error())
			return
		}
		ezap.Errorf("[socks5] connect to %s", net.JoinHostPort(host, port))
		client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //response to client connection is done.
		tunnel(client, server)
	} else if b[0] == 0x04 { //only for socks4
		var host, port string
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
		port = strconv.Itoa(int(b[2])<<8 | int(b[3]))
		server, err := st.Client.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			ezap.Error("[socks4] " + err.Error())
			return
		}
		ezap.Infof("[socks4] connect to %s", net.JoinHostPort(host, port))
		client.Write([]byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //response to client connection is done.
		tunnel(client, server)
	} else { //http
		s := string(b[:])
		ss := strings.Split(s, " ")
		method := ss[0]
		if method == "CONNECT" {
			host := ss[1]
			server, err := st.Client.Dial("tcp", host)
			if err != nil {
				ezap.Error("[http] " + err.Error())
				return
			}
			ezap.Infof("[http] connect to %s", host)
			success := []byte("HTTP/1.1 200 Connection established\r\n\r\n")
			_, err = client.Write(success)
			if err != nil {
				ezap.Error("[http] " + err.Error())
				return
			}
			tunnel(client, server)
		} else {
			u := ss[1]
			_url, _ := url.Parse(u)
			address := ""

			if !strings.Contains(_url.Host, ":") {
				if _url.Scheme == "http" {
					address = _url.Host + ":80"
				} else {
					address = _url.Host + ":443"
				}
			} else {
				address = _url.Host
			}

			server, err := st.Client.Dial("tcp", address)
			if err != nil {
				ezap.Error("[http] " + err.Error())
				return
			}
			ezap.Infof("[http] forward to %s", address)

			fmt.Fprint(server, s)

			tunnel(client, server)
		}
	}
}

func tunnel(client net.Conn, server net.Conn) {
	go func() {
		io.Copy(client, server)
		server.Close()
		client.Close()
	}()
	io.Copy(server, client)
	client.Close()
	server.Close()
}

func (st *SSHTun) Close() {
	if st.Client != nil {
		st.Client.Close()
	}
}
