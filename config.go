package combustion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"text/template"

	"github.com/coreos/ignition/config"
	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/ghodss/yaml"
	"github.com/src-d/combustion/transpiler"
	"github.com/vincent-petithory/dataurl"
	"srcd.works/go-billy.v1"
	"srcd.works/go-billy.v1/os"
)

var DefaultIgnitionVersion = types.IgnitionVersion{Major: 2}

// FileSystem used in any file operation
var FileSystem billy.Filesystem = os.New("")

type Config struct {
	Imports map[string]Values `json:"import,omitempty"`
	Output  string            `json:"output,omitempty"`
	Type    string            `json:"type,omitempty"`
	types.Config

	dir  string // dir where the config is located
	name string // config name
}

// NewConfigFromFile opens the given file and calls NewConfig with the given
// values
func NewConfigFromFile(filename string, values map[string]string) (*Config, error) {
	file, err := FileSystem.Open(filename)
	if err != nil {
		return nil, err
	}

	return NewConfig(file, filename, values)
}

// NewConfig returns a new Config unmashaling the r content interpolated with
// the given values. A dir, should be provided to be able to read and resolve
// all the includes
func NewConfig(r io.Reader, filename string, values map[string]string) (*Config, error) {
	c, err := newConfig(r, filename, values)
	if err != nil {
		return nil, err
	}

	return c, c.resolve()
}

func newConfig(r io.Reader, filename string, values map[string]string) (*Config, error) {
	c := &Config{}
	c.dir, c.name = filepath.Split(filename)

	if err := c.Unmarshal(r, values); err != nil {
		return nil, err
	}

	if c.Ignition.Version.Major == 0 {
		c.Ignition.Version = DefaultIgnitionVersion
	}

	return c, nil
}

// Unmarshal unmarshal the r content into Config, the content is interpolated
// using the given values
func (c *Config) Unmarshal(r io.Reader, values map[string]string) error {
	y, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	y, err = c.interpolate(y, values)
	if err != nil {
		return err
	}

	defer c.fixStorageFiles()
	return yaml.Unmarshal(y, c)
}

func (c *Config) interpolate(content []byte, v Values) ([]byte, error) {
	t, err := template.New("t").Parse(string(content))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (c *Config) fixStorageFiles() {
	// all the errors are ignored because if the url is real malformed will be
	// identified by the validator
	for i, f := range c.Storage.Files {
		s := f.Contents.Source
		u, err := dataurl.DecodeString(s.String())
		if err == nil || err.Error() != "missing data prefix" {
			continue
		}

		raw, err := url.QueryUnescape(s.String())
		if err != nil {
			continue
		}

		u = dataurl.New([]byte(raw), "text/plain")
		pu, _ := url.Parse(u.String())
		f.Contents.Source = types.Url(*pu)

		c.Storage.Files[i] = f
	}
}

func (c *Config) resolve() error {
	return c.doResolve(c.dir, make(stack, 0))
}

func (c *Config) doResolve(dir string, s stack) error {
	for file, values := range c.Imports {
		fullpath := filepath.Join(dir, string(file))
		if s.In(fullpath) {
			return &ErrCircularDependency{s, fullpath}
		}

		f, err := FileSystem.Open(fullpath)
		if err != nil {
			return err
		}

		src, err := newConfig(f, fullpath, values)
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

func (c *Config) append(src *Config) {
	c.Config = config.Append(c.Config, src.Config)
}

func (c *Config) SaveTo(dir string) (report.Report, error) {
	var r report.Report
	if c.Output == "" {
		return r, nil
	}

	fullpath := FileSystem.Join(dir, c.Output)
	file, err := FileSystem.Create(fullpath)
	if err != nil {
		return r, err
	}

	return c.Render(file)
}

func (c *Config) Render(w io.Writer) (r report.Report, err error) {
	var content []byte

	switch c.Type {
	case "cloud-config":
		content, r, err = c.marshalToCloudConfig()
	default:
		content, r, err = c.marshalToIgnition()
	}

	if err != nil {
		return r, err
	}

	_, err = w.Write(content)
	return r, err
}

func (c *Config) marshalToIgnition() ([]byte, report.Report, error) {
	var r report.Report

	json, err := json.MarshalIndent(c.Config, "", "  ")
	if err != nil {
		return nil, r, err
	}

	_, r, err = config.ParseFromLatest(json)
	if err != nil {
		return json, r, err
	}

	return json, r, nil
}

func (c *Config) marshalToCloudConfig() ([]byte, report.Report, error) {
	cc, r := transpiler.TranspileIgnition(&c.Config)
	y, err := marchalToYAML(cc)
	return y, r.Report, err
}

// Values interpolation values to replace on the Config
type Values map[string]string
type stack []string

func (s stack) In(filename string) bool {
	for _, f := range s {
		if f == filename {
			return true
		}
	}

	return false
}

type ErrCircularDependency struct {
	Stack stack
	File  string
}

func (err *ErrCircularDependency) Error() string {
	return fmt.Sprintf(
		"circular-dependency detected including %q: %s",
		err.File, err.Stack,
	)
}
