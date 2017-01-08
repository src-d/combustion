package combustion

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNetworkd(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"networkd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      contents: |\n" +
		"        [Foo]\n" +
		"        qux=qux\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/valid.yaml", nil)
	assert.NoError(t, err)

	r := c.validate()
	assert.Len(t, r.Entries, 0)
}

func TestValidateNetworkdEmpty(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"networkd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/empty.yaml", nil)
	assert.NoError(t, err)

	r := c.validate()
	assert.Len(t, r.Entries, 1)
}

func TestValidateNetworkdMalformed(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"networkd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      contents: |\n" +
		"        [Foo\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/malformed.yaml", nil)
	assert.NoError(t, err)

	r := c.validate()
	assert.Len(t, r.Entries, 1)
}

func TestValidateSystemd(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      contents: |\n" +
		"        [Foo]\n" +
		"        qux=qux\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/valid.yaml", nil)
	assert.NoError(t, err)

	r := c.validate()
	assert.Len(t, r.Entries, 0)
}

func TestValidateSystemdMalformed(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      contents: |\n" +
		"        [Foo\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/malformed.yaml", nil)
	assert.NoError(t, err)

	r := c.validate()
	assert.Len(t, r.Entries, 1)
}

func TestValidateSystemdEmpty(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/empty.yaml", nil)
	assert.NoError(t, err)

	r := c.validate()
	assert.Len(t, r.Entries, 1)
}

func TestValidateSystemdEmptyValidDropIn(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      dropins:\n" +
		"      - name: 50-insecure-registry.conf\n" +
		"        contents: |\n" +
		"          [Qux]\n" +
		"          foo=bar\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/empty-valid-dropin.yaml", nil)
	assert.NoError(t, err)

	r := c.validate()
	assert.Len(t, r.Entries, 0)
}

func TestValidateSystemdMalformedDropIn(t *testing.T) {
	input := []byte("" +
		"---\n" +
		"systemd:\n" +
		"  units:\n" +
		"    - name: installer.service\n" +
		"      dropins:\n" +
		"      - name: 50-insecure-registry.conf\n" +
		"        contents: |\n" +
		"          [Qux\n" +
		"",
	)

	c, err := NewConfig(bytes.NewBuffer(input), "fixtures/malformed-dropin.yaml", nil)
	assert.NoError(t, err)

	r := c.validate()
	assert.Len(t, r.Entries, 1)
}
