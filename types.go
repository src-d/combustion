package combustion

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/coreos/ignition/config"
	"github.com/coreos/ignition/config/types"
)

type Config struct {
	Include
	types.Config
}

func (c *Config) Resolve(base string) error {
	return c.doResolve(base, make(Stack, 0))
}

func (c *Config) doResolve(base string, s Stack) error {
	for file, _ := range c.File {
		fullpath := filepath.Join(base, string(file))
		if s.In(fullpath) {
			return &ErrCircularDependency{s, fullpath}
		}

		src, err := file.Load(base)
		if err != nil {
			return err
		}

		if err := src.doResolve(filepath.Dir(fullpath), append(s, fullpath)); err != nil {
			return err
		}

		c.append(src)
	}

	return nil
}

func (c *Config) append(src Config) {
	c.Config = config.Append(c.Config, src.Config)
}

type Include struct {
	File map[ConfigFile]map[string]string `json:"include,omitempty"`
}

type ConfigFile string

func (f ConfigFile) Load(base string) (Config, error) {
	var c Config

	fullpath := filepath.Join(base, string(f))
	y, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return c, err
	}

	return c, yaml.Unmarshal(y, &c)
}

type Stack []string

func (s Stack) In(filename string) bool {
	for _, f := range s {
		if f == filename {
			return true
		}
	}

	return false
}

type ErrCircularDependency struct {
	Stack Stack
	File  string
}

func (err *ErrCircularDependency) Error() string {
	return fmt.Sprintf(
		"circular-dependency detected including %q: %s",
		err.File, err.Stack,
	)
}
