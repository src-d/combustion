package combustion

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
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

func TestValidateMalformed(t *testing.T) {
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

func TestValidateEmpty(t *testing.T) {
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

func TestValidateEmptyValidDropIn(t *testing.T) {
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

func TestValidateMalformedDropIn(t *testing.T) {
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
