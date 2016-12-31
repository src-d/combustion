package combustion

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

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
	for file, values := range c.File {
		fullpath := filepath.Join(base, string(file))
		if s.In(fullpath) {
			return &ErrCircularDependency{s, fullpath}
		}

		src, err := file.Load(base, values)
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

func (f ConfigFile) Load(base string, values map[string]string) (Config, error) {
	var c Config

	fullpath := filepath.Join(base, string(f))
	y, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return c, err
	}

	y, err = f.interpolate(y, values)
	if err != nil {
		return c, err
	}

	return c, yaml.Unmarshal(y, &c)
}

func (f ConfigFile) interpolate(content []byte, values map[string]string) ([]byte, error) {
	if len(values) == 0 {
		return content, nil
	}

	t, err := template.New("t").Parse(string(content))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, values); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
