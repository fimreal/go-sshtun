package gosshtun

import (
	"github.com/fimreal/goutils/ezap"
	"github.com/wzshiming/sysproxy"
)

func (st *SSHTun) PacOn(rule string) (ok bool) {
	err := sysproxy.OnPAC(rule)
	if err != nil {
		ezap.Errorf("unable use pac rule[%s]: %s", rule, err)
		return
	}
	ezap.Info("enable proxy auto-config(pac): ", st.PacInspect())
	return true
}

func (st *SSHTun) PacInspect() string {
	out, err := sysproxy.GetPAC()
	if err != nil {
		ezap.Errorf("unable get pac rule: %s", err)
		return ""
	}
	return out
}

func (st *SSHTun) PacOff() {
	err := sysproxy.OffPAC()
	if err != nil {
		ezap.Errorf("unable prue pac rule: %s", err)
	}
	ezap.Info("disable proxy auto-config(pac)")
}
