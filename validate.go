package combustion

import (
	"bytes"
	"fmt"

	"github.com/coreos/go-systemd/unit"
	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
)

func (c *Config) validate() report.Report {
	var r report.Report

	checkNetworkdUnits(c.Config, &r)
	checkSystemdUnits(c.Config, &r)
	return r
}

func checkNetworkdUnits(cfg types.Config, r *report.Report) {
	for _, u := range cfg.Networkd.Units {
		if err := validateUnitContent(u.Contents); err != nil {
			r.Add(report.Entry{
				Kind:    report.EntryError,
				Message: err.Error(),
			})
		}
	}
}

func checkSystemdUnits(cfg types.Config, r *report.Report) {
	for _, u := range cfg.Systemd.Units {
		for _, err := range checkSystemdUnit(&u) {
			r.Add(report.Entry{
				Kind:    report.EntryError,
				Message: err.Error(),
			})
		}
	}
}

func checkSystemdUnit(u *types.SystemdUnit) []error {
	var errs []error

	if err := validateUnitContent(u.Contents); err != nil {
		if err != errEmptyUnit || (err == errEmptyUnit && len(u.DropIns) == 0) {
			errs = append(errs, fmt.Errorf("%s (unit: %q)", err, u.Name))
		}
	}

	for _, d := range u.DropIns {
		if err := validateUnitContent(d.Contents); err != nil {
			errs = append(errs, fmt.Errorf("%s (drop-in: %q)", err, d.Name))
		}
	}

	return errs
}

var errEmptyUnit = fmt.Errorf("invalid or empty unit content")

func validateUnitContent(content string) error {
	c := bytes.NewBufferString(content)
	unit, err := unit.Deserialize(c)
	if err != nil {
		return fmt.Errorf("invalid unit content: %s", err)
	}

	if len(unit) == 0 {
		return errEmptyUnit
	}

	return nil
}
