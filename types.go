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
	history := make(map[string]bool, 0)
	return c.doResolve(base, history)
}

func (c *Config) doResolve(base string, history map[string]bool) error {
	for file, vars := range c.File {
		fullpath := filepath.Join(base, string(file))
		fmt.Println(base, fullpath, vars)

		if history[fullpath] {
			return fmt.Errorf("loop %q", fullpath)
		}

		history[fullpath] = true

		src, err := file.Load(base)
		if err != nil {
			return err
		}

		if err := src.doResolve(filepath.Dir(fullpath), history); err != nil {
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
