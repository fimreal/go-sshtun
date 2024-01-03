package gosshtun

import (
	"encoding/base64"
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
		ezap.Errorf("not figure out protocol: %s", err)
		client.Close()
		return
	}
	// get client ip
	clientIP := strings.Split(client.RemoteAddr().String(), ":")[0]

	if b[0] == 0x05 { // for socks5
		ezap.Debug(b[:])
		// need to validation
		if st.Auth != "" {
			if int(b[1]) < 2 {
				ezap.Errorf("[socks5] %s No Acceptable Methods", clientIP)
				// no username/password => reject
				client.Write([]byte{b[0], 0xff})
				client.Close()
				return
			}
			client.Write([]byte{b[0], 0x02})
			_, err := client.Read(b[:])
			if err != nil {
				ezap.Errorf("[socks5] %s not find remote address: %s", clientIP, err)
				client.Close()
				return
			}
			ezap.Debug(b[:])

			usernameLens := int(b[1])
			username := string(b[2 : 2+usernameLens])
			passwordLens := int(b[2+usernameLens])
			password := string(b[2+usernameLens+1 : 2+usernameLens+1+passwordLens])
			authStr := username + ":" + password
			ezap.Debugf("[socks5] %s get authstr: ", clientIP, authStr)
			if authStr != st.Auth {
				ezap.Errorf("[socks5] %s remote no valid auth", clientIP)
				// wrong username/password => reject
				client.Write([]byte{b[0], 0xff})
				client.Close()
				return
			}
		}
		// response to client: no need to validation
		client.Write([]byte{b[0], 0x00})

		n, err := client.Read(b[:])
		if err != nil {
			ezap.Errorf("[socks5] %s not find remote address: %s", clientIP, err)
			client.Close()
			return
		}

		ezap.Debug(b[:])

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
			client.Close()
			ezap.Fatal("[socks5] %s fail to dial the host", clientIP, err)
		}
		ezap.Infof("[socks5] %s connect to %s", clientIP, net.JoinHostPort(host, port))
		client.Write([]byte{b[0], 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //response to client connection is done.
		st.tunnel(client, server)
	} else if b[0] == 0x04 {
		// for socks4
		if st.Auth != "" {
			ezap.Errorf("[socks4] %s connection was rejected, because the server has enabled authentication", clientIP)
			client.Write([]byte{0x5B})
			client.Close()
			return
		}
		var host, port string
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
		port = strconv.Itoa(int(b[2])<<8 | int(b[3]))
		server, err := st.Client.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			client.Close()
			ezap.Fatalf("[socks4] %s fail to dial the host", clientIP, err)
		}
		ezap.Infof("[socks4] %s connect to %s", clientIP, net.JoinHostPort(host, port))
		client.Write([]byte{0x00, 0x5a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //response to client connection is done.
		st.tunnel(client, server)
	} else {
		// for http
		s := string(b[:])
		ezap.Debug(s)
		ss := strings.Split(s, " ")
		if len(ss) < 2 {
			ezap.Errorf("[http] %s request seems to be invalid: %s", clientIP, s)
			return
		}
		method := ss[0]
		host := ss[1]
		// ezap.Debugf("%+v", s)
		if method == "CONNECT" {
			server, err := st.Client.Dial("tcp", host)
			if err != nil {
				client.Close()
				ezap.Fatal("[http] %s fail to dial the host: %s", clientIP, err)
			}
			ezap.Infof("[http] %s connect to %s", clientIP, host)
			success := []byte("HTTP/1.1 200 Connection established\r\n\r\n")
			_, err = client.Write(success)
			if err != nil {
				ezap.Errorf("[http] %s connect to %s", clientIP, host)
				client.Close()
				return
			}
			st.tunnel(client, server)
		} else {
			_url, err := url.Parse(host)
			if err != nil {
				ezap.Errorf("[http] %s could not parse url: %s", clientIP, err)
				client.Close()
				return
			}
			if _url.RawQuery == "rs=sshtun" {
				// build-in http handle
				if strings.HasSuffix(_url.Path, ".pac") {
					// for embeded pac rule
					defer client.Close()
					filename := strings.TrimPrefix(_url.Path, "/pac/")
					ezap.Infof("[in] %s request pac file: ", clientIP, filename)
					f, err := pacFiles.Open(filename)
					if err != nil {
						ezap.Errorf("[in] fail to open file: %s", err)
						_, err = client.Write([]byte("HTTP/1.1 404\r\n\r\n"))
						if err != nil {
							ezap.Errorf("[in] %s fail to reply 404 to the connection: %s", clientIP, err)
						} else {
							ezap.Infof("[in] %s return 404", clientIP)
						}
						return
					}
					defer f.Close()
					proxy, err := proxyAddr(st.ListenAddr)
					if err != nil {
						ezap.Fatal("Could not parse listen address: ", err)
					}
					_, err = client.Write([]byte("HTTP/1.1 200\r\n\r\n"))
					if err != nil {
						ezap.Errorf("[in] %s err handle request: %s", clientIP, err)
						return
					}
					_, err = client.Write([]byte("var  proxy = \"SOCKS5 " + proxy + "; SOCKS " + proxy + "; PROXY " + proxy + "; DIRECT;\";\r\n\r\n"))
					if err != nil {
						ezap.Errorf("[in] %s fail to write data to the connection: %s", clientIP, err)
						return
					}
					_, err = io.Copy(client, f)
					if err != nil {
						ezap.Errorf("[in] %s err: handle request: %s", clientIP, err)
						return
					}
					return
				} else if _url.Path == "/stat" {
					// for traffic statistics
					defer client.Close()
					client.Write([]byte("HTTP/1.1 200\r\n\r\n"))
					client.Write([]byte("SSH Tunnel:\r\nServer: " + st.Client.RemoteAddr().String() + "\r\nListen Address: " + st.ListenAddr))
					client.Write([]byte("\r\n\r\nStatistic:\r\nTotal Upload: " + beautifySize(atomic.LoadInt64(&st.TotalUpload)) + "\r\nTotal Download: " + beautifySize(atomic.LoadInt64(&st.TotalDownload))))
					return
				}
			} else {
				if st.Auth != "" {
					bs64Auth := base64.StdEncoding.EncodeToString([]byte(st.Auth))
					if !strings.Contains(s, "Proxy-Authorization: Basic "+bs64Auth+"\r\n") {
						ezap.Errorf("[http] %s connection was rejected, because the server has enabled authentication and no valid user/password", clientIP)
						client.Write([]byte("HTTP/1.1 401 Unauthorized\r\n\r\n"))
						client.Close()
						return
					}
				}
				// forward
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
					ezap.Errorf("[http] %s err: %s", clientIP, err)
					client.Close()
					return
				}

				ezap.Infof("[http] %s forward to %s", clientIP, address)
				fmt.Fprint(server, s)
				st.tunnel(client, server)
			}
		}
	}
}

func (st *SSHTun) tunnel(client net.Conn, server net.Conn) {
	go func() {
		defer server.Close()
		defer client.Close()
		n, _ := io.Copy(client, server)
		atomic.AddInt64(&st.TotalUpload, n)
		// ezap.Debug("TotalUpload: ", atomic.LoadInt64(&st.TotalUpload))
	}()
	defer client.Close()
	defer server.Close()
	n, _ := io.Copy(server, client)
	atomic.AddInt64(&st.TotalDownload, n)
	// ezap.Debug("TotalDownload: ", atomic.LoadInt64(&st.TotalDownload))
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
