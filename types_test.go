package combustion

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	input := []byte(`---
systemd:
  units:
    - name: installer.service
      enable: true
      contents: |
        [Unit]
        Requires=network-online.target
        After=network-online.target
        [Service]
        Type=simple
        ExecStart=/opt/installer
        [Install]
        WantedBy=multi-user.target
    `)

	var c Config
	err := yaml.Unmarshal(input, &c)
	assert.NoError(t, err)

	js, err := json.MarshalIndent(c, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(js))
}

func TestConfigResolve(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"include:\n" +
		"  include/foo.yaml:\n" +
		"  include/bar.yaml:\n" +
		"",
	)

	var c Config
	err := yaml.Unmarshal(input, &c)
	assert.NoError(t, err)

	err = c.Resolve("fixtures")
	assert.NoError(t, err)

	var names []string
	for _, u := range c.Config.Systemd.Units {
		names = append(names, string(u.Name))
	}

	sort.Strings(names)
	assert.EqualValues(t, []string{"bar", "baz", "foo", "qux", "qux"}, names)
}

func TestConfigResolveCircular(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"include:\n" +
		"  circular/foo.yaml:\n" +
		"",
	)

	var c Config
	err := yaml.Unmarshal(input, &c)
	assert.NoError(t, err)

	err = c.Resolve("fixtures")
	assert.IsType(t, &ErrCircularDependency{}, err)

}

func TestConfigResolveInterpolate(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"include:\n" +
		"  interpolate/foo.yaml:\n" +
		"    foo: bar\n" +
		"",
	)

	var c Config
	err := yaml.Unmarshal(input, &c)
	assert.NoError(t, err)

	err = c.Resolve("fixtures")
	assert.NoError(t, err)

	var names []string
	for _, u := range c.Config.Systemd.Units {
		names = append(names, string(u.Name))
	}

	assert.EqualValues(t, []string{"bar"}, names)

}
