"use strict";

var direct = "DIRECT"
var proxy = "SOCKS5 127.0.0.1:1080; SOCKS 127.0.0.1:1080; PROXY 127.0.0.1:1080;" + direct;

var ProxyHostList = {
    "youtube.com": true,
};

function FindProxyForURL(url, host) {
    // proxy list
    if (host.indexOf(".") < 0
        || ProxyHostList[host]
        || /\.?google/.test(host)
    ) {
        return proxy;
    }

    // if (/^(\d{1,3}\.){3}\d{1,3}$/.test(host)
    //     && (isInNet(host, "127.0.0.0", "255.255.255.0")
    //         || isInNet(host, "192.168.0.0", "255.255.0.0")
    //         || isInNet(host, "172.16.0.0", "255.240.0.0")
    //         || isInNet(host, "10.0.0.0", "255.0.0.0")
    //         || isInNet(host, "202.113.16.0", "255.255.240.0")
    //         || isInNet(host, "202.113.224.0", "255.255.240.0")
    //         || isInNet(host, "222.30.61.0", "255.255.225.0"))
    // ) {
    //     return direct;
    // }

    return direct;
};