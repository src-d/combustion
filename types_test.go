package combustion

import (
	"encoding/json"
	"fmt"
	"testing"

	"sort"

	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/ghodss/yaml"
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
	assert.NilError(t, err)

	js, err := json.MarshalIndent(c, "", "  ")
	assert.NilError(t, err)
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
	assert.NilError(t, err)

	err = c.Resolve("fixtures")
	assert.NilError(t, err)

	var names []string
	for _, u := range c.Config.Systemd.Units {
		names = append(names, string(u.Name))
	}

	sort.Strings(names)
	assert.DeepEqual(t, names, []string{"bar", "baz", "foo", "qux"})
}
