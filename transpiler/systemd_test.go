package transpiler

import (
	"net/url"
	"testing"

	"github.com/coreos/ignition/config/types"
	"github.com/docker/docker/pkg/testutil/assert"
)

func TestDoSystemd(t *testing.T) {
	url, _ := url.Parse("data:,foo")
	c := &types.Config{}
	c.Storage.Files = []types.File{{
		Node: types.Node{
			Filesystem: "root",
			Path:       "/foo/bar",
			Mode:       types.NodeMode(420),
		},
		Contents: types.FileContents{
			Source: types.Url(*url),
		},
	}}

	cc, _ := TranspileIgnition(c)

	file := cc.WriteFiles[0]
	assert.Equal(t, "/foo/bar", file.Path)
	assert.Equal(t, "foo", file.Content)
	assert.Equal(t, "", file.Owner)
	assert.Equal(t, "0644", file.RawFilePermissions)
}
