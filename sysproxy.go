package gosshtun

import (
	"github.com/fimreal/goutils/ezap"
	"github.com/wzshiming/sysproxy"
)

// ref. https://github.com/wzshiming/sysproxy/blob/master/cmd/sysproxy/main.go
func (st *SSHTun) EnableSystemProxy() (ok bool) {
	proxy, err := proxyAddr(st.ListenAddr)
	if err != nil {
		ezap.Fatalf("listen address[%s] is not vaild: %s", st.ListenAddr, err)
	}

	// set noproxy
	err = sysproxy.OnNoProxy([]string{"127.0.0.1", "localhost"})
	if err != nil {
		ezap.Errorf("error set system noproxy: %s", err)
		return
	}

	// set http proxy
	err = sysproxy.OnHTTP(proxy)
	if err != nil {
		ezap.Errorf("error set system http proxy: %s", err)
		return
	}
	out, err := sysproxy.GetHTTP()
	if err != nil {
		ezap.Errorf("error get system http proxy: %s", err)
		return
	}
	ezap.Info("set system http proxy: ", out)

	// set https proxy
	err = sysproxy.OnHTTPS(proxy)
	if err != nil {
		ezap.Errorf("error set system https proxy: %s", err)
		return
	}
	out, err = sysproxy.GetHTTPS()
	if err != nil {
		ezap.Errorf("error get system https proxy: %s", err)
		return
	}
	ezap.Info("set system https proxy: ", out)
	return true
}

func DisableSystemProxy() {
	err := sysproxy.OffHTTP()
	if err != nil {
		ezap.Errorf("unable prue pac rule: %s", err)
	} else {
		ezap.Info("unset system https proxy")
	}

	err = sysproxy.OffHTTPS()
	if err != nil {
		ezap.Debug(err.Error())
	} else {
		ezap.Info("unset system https proxy")
	}
}
