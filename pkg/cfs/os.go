package cfs

import (
	"io/fs"
	"os"
)

// OS is a fs.FS implementation that reads files from the host OS.
type OS struct{}

var _ fs.FS = (*OS)(nil) // ensure interface is implemented

// Open implements fs.FS.
func (*OS) Open(name string) (fs.File, error) {
	return os.Open(name)
}
