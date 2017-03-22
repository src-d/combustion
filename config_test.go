package combustion

import (
	"bytes"
	"sort"
	"testing"

	"github.com/coreos/container-linux-config-transpiler/config/types"
	cc "github.com/coreos/coreos-cloudinit/config"
	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-billy.v2/memfs"
	yaml "gopkg.in/yaml.v1"
)

func init() {
	FileSystem = memfs.New()
}

func TestNewConfig(t *testing.T) {
	WriteFixtures([][]string{{
		"fixtures/import/foo.yaml",
		"import:\n    baz.yaml:\n\nsystemd:\n  units:\n   - name: foo\n  conent: foo\n",
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
	WriteFixture("fixtures/interpolate/foo.yaml", "systemd:\n  units:\n    - name: {{.foo}} {% .foo %} ")

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

	assert.EqualValues(t, []string{"bar {{ .foo }}"}, names)
}

func TestConfigResolveInterpolateMissing(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: {{.foo}}\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/inline.yaml", nil)
	assert.Error(t, err)
	assert.Nil(t, c)
}

func TestConfigFixStorageFiles(t *testing.T) {
	WriteFixture("fixtures/foo.txt", "bar")
	input := []byte("" +
		"---\n" +
		"storage:\n" +
		"  files:\n" +
		"    - path: test\n" +
		"      contents:\n" +
		"        remote: \n" +
		"          url: |\n" +
		"            file:///foo.txt" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/inline.yaml", nil)
	assert.NoError(t, err)

	assert.Equal(t, "", c.Storage.Files[0].Contents.Remote.Url)
	assert.Equal(t, "bar", c.Storage.Files[0].Contents.Inline)
}

func TestConfigRender(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      enable: true\n" +
		"      contents: |\n" +
		"        [foo]\n" +
		"        foo=bar\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/inline.yaml", nil)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(nil)
	r, err := c.Render(buf)
	assert.NoError(t, err)
	assert.Equal(t, r.IsFatal(), false)

	var result types.Config
	err = yaml.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)

	//assert.Equal(t, DefaultIgnitionVersion, result.Ignition.Version)
	assert.Equal(t, 1, len(result.Systemd.Units))
}

func TestConfigRenderToCloudConfig(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"type: cloud-config\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      enable: true\n" +
		"      contents: |\n" +
		"        [foo]\n" +
		"        foo=bar\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/inline.yaml", nil)
	assert.NoError(t, err)

	buf := bytes.NewBuffer(nil)
	r, err := c.Render(buf)
	assert.NoError(t, err)
	assert.Equal(t, r.IsFatal(), false)

	var result cc.CloudConfig
	err = yaml.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(result.CoreOS.Units))
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
