package transpiler

import (
	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/ignition/config/types"
)

func (t *ccTranspiler) doSystemd(s *types.Systemd) {
	for i, u := range s.Units {
		t.doSystemdUnit(i, &u)
	}
}

func (t *ccTranspiler) doSystemdUnit(idx int, u *types.SystemdUnit) {
	t.cc.CoreOS.Units = append(t.cc.CoreOS.Units, config.Unit{
		Name:    string(u.Name),
		Enable:  u.Enable,
		Mask:    u.Mask,
		Runtime: true,
		Content: u.Contents,
	})
}
