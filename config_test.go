package combustion

import (
	"bytes"
	"encoding/json"
	"sort"
	"testing"

	"github.com/coreos/ignition/config/types"
	"github.com/stretchr/testify/assert"
	"srcd.works/go-billy.v1/memory"
)

func init() {
	FileSystem = memory.New()
}

func TestNewConfig(t *testing.T) {
	WriteFixtures([][]string{{
		"fixtures/import/foo.yaml",
		"import:\n    baz.yaml:\n\nsystemd:\n  units:\n   - name: foo",
	}, {
		"fixtures/import/baz.yaml",
		"import:\n    folder/qux.yaml:\n\nsystemd:\n  units:\n   - name: baz",
	}, {
		"fixtures/import/bar.yaml",
		"import:\n    folder/qux.yaml:\n\nsystemd:\n  units:\n   - name: bar",
	}, {
		"fixtures/import/folder/qux.yaml",
		"systemd:\n  units:\n   - name: qux",
	}})

	input := []byte("" +
		"---\n" +
		"import:\n" +
		"  import/foo.yaml:\n" +
		"  import/bar.yaml:\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/inline.yaml", nil)
	assert.NoError(t, err)

	var names []string
	for _, u := range c.Config.Systemd.Units {
		names = append(names, string(u.Name))
	}

	sort.Strings(names)
	assert.EqualValues(t, []string{"bar", "baz", "foo", "qux", "qux"}, names)
}

func TestConfigResolveCircular(t *testing.T) {
	WriteFixture("fixtures/circular/foo.yaml", "import:\n    bar.yaml:")
	WriteFixture("fixtures/circular/bar.yaml", "import:\n    foo.yaml:")

	input := []byte("" +
		"---\n" +
		"import:\n" +
		"  circular/foo.yaml:\n" +
		"",
	)

	_, err := NewConfig(bytes.NewBuffer(input), "fixtures/inline.yaml", nil)
	assert.IsType(t, &ErrCircularDependency{}, err)
}

func TestConfigResolveInterpolate(t *testing.T) {
	WriteFixture("fixtures/interpolate/foo.yaml", "systemd:\n  units:\n    - name: {{.foo}}")

	input := []byte("" +
		"---\n" +
		"import:\n" +
		"  interpolate/foo.yaml:\n" +
		"    foo: bar\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/inline.yaml", nil)
	assert.NoError(t, err)

	var names []string
	for _, u := range c.Config.Systemd.Units {
		names = append(names, string(u.Name))
	}

	assert.EqualValues(t, []string{"bar"}, names)
}

func TestConfigRender(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      enable: true\n" +
		"      contents: foo\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/inline.yaml", nil)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(nil)
	r, err := c.Render(buf)
	assert.Equal(t, r.IsFatal(), false)
	assert.NoError(t, err)

	var result types.Config
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, DefaultIgnitionVersion, result.Ignition.Version)
	assert.Equal(t, 1, len(result.Systemd.Units))
}

func WriteFixtures(fixtures [][]string) {
	for _, f := range fixtures {
		WriteFixture(f[0], f[1])
	}
}

func WriteFixture(path, content string) {
	f, err := FileSystem.Create(path)
	if err != nil {
		panic(err)
	}

	if _, err := f.Write([]byte(content)); err != nil {
		panic(err)
	}
}
