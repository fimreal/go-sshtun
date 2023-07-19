# go-sshtun
ssh tunnel via http, socks, support set pac or global system proxy

# useage
#### quick start
```bash
docker run --rm --network host epurs/sshtun:lastest
```

#### download from release 

```bash
# sshtun --help
  -d, --debug                   debug mode
  -h, --host string             ssh server address, like "192.168.50.220:2222"
  -i, --identitykey string      identity key file
  -k, --identitykeydir string   identity key dir
  -l, --listen string           service listing on (default "127.0.0.1:1080")
  -P, --pac string              enable pac. need a pac url, like "http://www.example.com/my.pac", or use embedded rules(gfw, tiny)
  -p, --password string         ssh password
  -S, --sysproxy                enable system proxy
  -t, --timeout duration        timeout (default 10s)
  -u, --user string             ssh user
```

#### docker run

```bash
# docker run --rm epurs/sshtun:lastest --help

# docker run --rm --network host -e "USER=root" -e "HOST=epurs.com" -e "PASSWORD=123456" -E "LISTEN=0.0.0.0:1080" epurs/sshtun:lastest

docker run -d --name sshtun1080 \
--restart unless-stopped \
-p 1080:1080 \
-v /Users/fimreal/.ssh:/root/.ssh \
epurs/sshtun:lastest \
-h epurs.com \
-uroot \
-i /root/.ssh/id_ed25519 \
-l 0.0.0.0:1080
```

# reference

https://github.com/ejunjsh/goproxy
https://github.com/scchenyong/sshtunnel/