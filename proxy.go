package gosshtun

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"

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
	if b[0] == 0x05 {
		// for socks5
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
			ezap.Errorf("[socks5] ", err)
			return
		}
		ezap.Infof("[socks5] connect to %s", net.JoinHostPort(host, port))
		client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //response to client connection is done.
		st.tunnel(client, server)
	} else if b[0] == 0x04 {
		// for socks4
		var host, port string
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
		port = strconv.Itoa(int(b[2])<<8 | int(b[3]))
		server, err := st.Client.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			ezap.Errorf("[socks4] ", err)
			return
		}
		ezap.Infof("[socks4] connect to %s", net.JoinHostPort(host, port))
		client.Write([]byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //response to client connection is done.
		st.tunnel(client, server)
	} else {
		// for http
		s := string(b[:])
		ss := strings.Split(s, " ")
		method := ss[0]
		host := ss[1]
		_url, err := url.Parse(host)
		if err != nil {
			ezap.Errorf("[http] could not parse url: %s", err)
			return
		}
		if method == "CONNECT" {
			server, err := st.Client.Dial("tcp", host)
			if err != nil {
				ezap.Errorf("[http] %s", err)
				return
			}
			ezap.Infof("[http] connect to %s", host)
			success := []byte("HTTP/1.1 200 Connection established\r\n\r\n")
			_, err = client.Write(success)
			if err != nil {
				ezap.Errorf("[http] connect to %s", host)
				return
			}
			st.tunnel(client, server)
		} else if _url.RawQuery == "rs=sshtun" {
			if strings.HasSuffix(_url.Path, ".pac") {
				// for embeded pac rule
				defer client.Close()
				filename := strings.TrimPrefix(_url.Path, "/pac/")
				ezap.Info("[in] request pac file: ", filename)
				f, err := pacFiles.Open(filename)
				if err != nil {
					ezap.Errorf("[in] fail to open file: %s", err)
					_, err = client.Write([]byte("HTTP/1.1 404\r\n\r\n"))
					if err != nil {
						ezap.Errorf("[in] fail to reply 404 to the connection: %s", err)
					} else {
						ezap.Info("[in] return 404")
					}
					return
				}
				defer f.Close()
				proxy, _ := proxyAddr(st.ListenAddr)
				_, err = client.Write([]byte("HTTP/1.1 200\r\n\r\n"))
				if err != nil {
					ezap.Errorf("[in] err handle request: %s", err)
					return
				}
				_, err = client.Write([]byte("var  proxy = \"SOCKS5 " + proxy + "; SOCKS " + proxy + "; PROXY " + proxy + "; DIRECT;\";\r\n\r\n"))
				if err != nil {
					ezap.Errorf("[in] fail to write data to the connection: %s", err)
					return
				}
				_, err = io.Copy(client, f)
				if err != nil {
					ezap.Errorf("[in] err: handle request: %s", err)
					return
				}
				return
			} else if _url.Path == "/stat" {
				// for traffic statistics
				defer client.Close()
				client.Write([]byte("HTTP/1.1 200\r\n\r\n"))
				client.Write([]byte("SSH Tunnel:\r\nServer: " + st.Client.RemoteAddr().String() + "\r\nListen Address: " + st.ListenAddr))
				client.Write([]byte("\r\n\r\nStatistic:\r\nTotal Upload: " + beautifySize(st.TotalUpload) + "\r\nTotal Download: " + beautifySize(st.TotalDownload)))
				return
			}
		} else {
			var address string
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
				ezap.Errorf("[http] ", err)
				return
			}

			ezap.Infof("[http] forward to %s", address)
			fmt.Fprint(server, s)
			st.tunnel(client, server)
		}
	}
}

func (st *SSHTun) tunnel(client net.Conn, server net.Conn) {
	go func() {
		n, _ := io.Copy(client, server)
		server.Close()
		client.Close()
		atomic.AddInt64(&st.TotalUpload, n)
	}()
	n, _ := io.Copy(server, client)
	client.Close()
	server.Close()
	atomic.AddInt64(&st.TotalDownload, n)
}

// 获取可访问地址
func proxyAddr(listenAddr string) (proxy string, err error) {
	_url, err := url.Parse("http://" + listenAddr)
	if err != nil {
		return
	}
	if strings.HasPrefix(_url.Host, "0.0.0.0:") {
		proxy = "127.0.0.1:" + _url.Port()
	} else {
		proxy = _url.Host
	}
	return
}
