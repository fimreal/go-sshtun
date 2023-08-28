package gosshtun

import (
	"embed"
	"io/fs"

	"github.com/fimreal/goutils/ezap"
	"github.com/wzshiming/sysproxy"
)

var (
	//go:embed pac
	embedPAC    embed.FS
	pacFiles, _ = fs.Sub(embedPAC, "pac")
)

func (st *SSHTun) PacOn(rule string) (ok bool) {
	proxy, _ := proxyAddr(st.ListenAddr)
	if rule == "gfw" || rule == "tiny" {
		rule = "http://" + proxy + "/pac/" + rule + ".pac?rs=sshtun"
	}
	err := sysproxy.OnPAC(rule)
	if err != nil {
		ezap.Errorf("unable use pac rule[%s]: %s", rule, err)
		return
	}
	ezap.Info("enable proxy auto-config(pac): ", PacInspect())
	return true
}

func PacInspect() string {
	out, err := sysproxy.GetPAC()
	if err != nil {
		ezap.Errorf("unable get pac rule: %s", err)
		return ""
	}
	return out
}

func PacOff() {
	err := sysproxy.OffPAC()
	if err != nil {
		ezap.Errorf("unable prue pac rule: %s", err)
	} else {
		ezap.Info("disable proxy auto-config(pac)")
	}
}
