# go-sshtun
ssh tunnel via http, socks

```bash
# sshtun --help
Usage of sshtun:
  -d, --debug                   debug mode
  -h, --host string             ssh server address, like "epurs.com:2222"
  -i, --identitykey string      identity key file
  -k, --identitykeydir string   identity key dir
  -l, --listen string           service listing on (default "127.0.0.1:10080")
  -p, --password string         ssh password
  -s, --sysproxy                enable system proxy
  -t, --timeout duration        timeout (default 10s)
  -u, --user string             ssh user
```

# reference

https://github.com/ejunjsh/goproxy
https://github.com/scchenyong/sshtunnel/