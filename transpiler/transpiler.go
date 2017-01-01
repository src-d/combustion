package transpiler

import (
	"fmt"
	"reflect"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
)

type Report struct {
	report.Report
}

func TranspileIgnition(c *types.Config) (*config.CloudConfig, *Report) {
	cc := &config.CloudConfig{}
	r := &Report{}

	t := &ccTranspiler{cc, r}
	t.TranspileIgnition(c)

	return cc, r
}

type ccTranspiler struct {
	cc *config.CloudConfig
	r  *Report
}

func (t *ccTranspiler) TranspileIgnition(c *types.Config) {
	t.doStorage(&c.Storage)
}

func IsZero(value interface{}) bool {
	zero := reflect.Zero(reflect.TypeOf(value)).Interface()
	if value != nil && !reflect.DeepEqual(value, zero) {
		return false
	}

	return true
}

func (t *ccTranspiler) ignoredEntry(key, reason string) {
	var msg string
	if reason == "" {
		msg = fmt.Sprintf("ignored %s, not supported in cloud-config", key)
	} else {
		msg = fmt.Sprintf("ignored %s, not supported in cloud-config, %s", key, reason)
	}

	t.r.Add(report.Entry{
		Kind:    report.EntryWarning,
		Message: msg,
	})
}
