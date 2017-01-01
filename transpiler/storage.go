package transpiler

import (
	"fmt"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/vincent-petithory/dataurl"
)

func (t *ccTranspiler) doStorage(s *types.Storage) {
	if !IsZero(s.Arrays) {
		t.r.Add(report.Entry{
			Kind:    report.EntryWarning,
			Message: "storage.arrays is not supported in cloud-config",
		})
	}

	if !IsZero(s.Disks) {
		t.r.Add(report.Entry{
			Kind:    report.EntryWarning,
			Message: "storage.disks is not supported in cloud-config",
		})
	}

	if !IsZero(s.Filesystems) {
		t.r.Add(report.Entry{
			Kind:    report.EntryWarning,
			Message: "storage.filesystems is not supported in cloud-config",
		})
	}

	t.doStorageFiles(s)
}

func (t *ccTranspiler) doStorageFiles(s *types.Storage) {
	for i, f := range s.Files {
		t.doStorageFile(i, &f)
	}
}

func (t *ccTranspiler) doStorageFile(idx int, f *types.File) {
	if f.Contents.Source.Scheme != "" && f.Contents.Source.Scheme != "data" {
		t.ignoredEntry(
			fmt.Sprintf("storage.files[%d]", idx),
			"only 'data' source are supported",
		)

		return
	}

	var owner string
	if !IsZero(f.User) && !IsZero(f.Group) {
		owner = fmt.Sprintf("%d:%d", f.User.Id, f.Group.Id)
	} else if !IsZero(f.User) {
		owner = fmt.Sprintf("%d", f.User.Id)
	}

	var content string
	if f.Contents.Source.Scheme != "" {
		url, err := dataurl.DecodeString(f.Contents.Source.String())
		if err != nil {
			fmt.Println(err)
			return
		}

		content = string(url.Data)
	}

	t.cc.WriteFiles = append(t.cc.WriteFiles, config.File{
		Content:            content,
		Path:               string(f.Path),
		Owner:              owner,
		RawFilePermissions: fmt.Sprintf("%#o", f.Mode),
	})

}
