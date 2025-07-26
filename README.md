# go-sshtun
ssh tunnel via http, socks, support set pac or global system proxy

# usage
#### quick start
```bash
docker run --rm --network host epurs/sshtun:latest
```

#### download from release 

```bash
# sshtun --help
  -a, --auth string             enable and set socks5/http proxy service authentication, eg. "user:pass"
  -d, --debug                   debug mode
  -h, --host string             ssh server address, eg. "192.168.50.220:2222"
  -i, --identitykey string      identity key file
  -k, --identitykeydir string   identity key dir
  -l, --listen string           service listing on (default "127.0.0.1:1080")
  -P, --pac string              enable pac (proxy auto-config). need a pac url, like "http://127.0.0.1:8000/my.pac", or use embedded rules(gfw, tiny)
  -p, --password string         ssh password
  -R, --reset                   reset/prune system proxy rule.
  -S, --sysproxy                enable system proxy
  -t, --timeout duration        timeout (default 10s)
  -u, --username string         ssh user
```

#### docker run

```bash
# docker run --rm epurs/sshtun:latest --help

# docker run --rm --network host -e "USER=root" -e "HOST=epurs.com" -e "PASSWORD=123456" -e "LISTEN=0.0.0.0:1080" epurs/sshtun:latest

# docker run --rm --network host -e "USER=sshuser" -e "HOST=sshhost" -e "PASSWORD=sshpassword" -e "LISTEN=0.0.0.0:1080" -e "AUTH=user:password" epurs/sshtun:latest

docker run -d --name sshtun1080 \
--restart unless-stopped \
-p 1080:1080 \
-v /Users/fimreal/.ssh:/root/.ssh \
epurs/sshtun:latest \
-h epurs.com \
-uroot \
-i /root/.ssh/id_ed25519 \
-l 0.0.0.0:1080
```

# reference

https://github.com/ejunjsh/goproxy

https://github.com/scchenyong/sshtunnel/
