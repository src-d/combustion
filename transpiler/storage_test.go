package transpiler

import (
	"net/url"
	"testing"

	"github.com/coreos/ignition/config/types"
	"github.com/docker/docker/pkg/testutil/assert"
)

func TestDoStorage(t *testing.T) {
	url, _ := url.Parse("data:,foo")
	c := &types.Config{}
	c.Storage.Files = []types.File{{
		Filesystem: "root",
		Path:       "/foo/bar",
		Contents: types.FileContents{
			Source: types.Url(*url),
		},
		Mode: types.FileMode(420),
	}}

	cc, _ := TranspileIgnition(c)

	file := cc.WriteFiles[0]
	assert.Equal(t, "/foo/bar", file.Path)
	assert.Equal(t, "foo", file.Content)
	assert.Equal(t, "", file.Owner)
	assert.Equal(t, "0644", file.RawFilePermissions)
}

func TestDoStorageUserAndGroup(t *testing.T) {
	c := &types.Config{}
	c.Storage.Files = []types.File{{
		User: types.FileUser{Id: 42},
	}, {
		User:  types.FileUser{Id: 42},
		Group: types.FileGroup{Id: 84},
	}}

	cc, _ := TranspileIgnition(c)
	assert.Equal(t, "42", cc.WriteFiles[0].Owner)
	assert.Equal(t, "42:84", cc.WriteFiles[1].Owner)
}

func TestDoStorageRaid(t *testing.T) {
	c := &types.Config{}
	c.Storage.Arrays = []types.Raid{{}}

	cc, r := TranspileIgnition(c)
	assert.NotNil(t, cc)
	assert.Equal(t, "storage.arrays is not supported in cloud-config", r.Entries[0].Message)
}

func TestDoStorageDisks(t *testing.T) {
	c := &types.Config{}
	c.Storage.Disks = []types.Disk{{}}

	cc, r := TranspileIgnition(c)
	assert.NotNil(t, cc)
	assert.Equal(t, "storage.disks is not supported in cloud-config", r.Entries[0].Message)
}

func TestDoStorageFilesystems(t *testing.T) {
	c := &types.Config{}
	c.Storage.Filesystems = []types.Filesystem{{}}

	cc, r := TranspileIgnition(c)
	assert.NotNil(t, cc)
	assert.Equal(t, "storage.filesystems is not supported in cloud-config", r.Entries[0].Message)
}
