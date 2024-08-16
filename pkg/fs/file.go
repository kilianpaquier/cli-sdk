package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Option represents a function taking an opt client to use filesysem package functions.
type Option func(opt option) option

// Join represents a function to join multiple elements between them.
type Join func(elems ...string) string

// WithJoin specifies a specific function to join a srcdir with one of its files in CopyDir.
func WithJoin(join Join) Option {
	return func(o option) option {
		o.join = join
		return o
	}
}

// WithPerm specifies the permission for target file for CopyFile and CopyDir.
func WithPerm(perm os.FileMode) Option {
	return func(o option) option {
		o.perm = perm
		return o
	}
}

// WithFS specifies a FS to read files instead of os filesystem.
func WithFS(fsys FS) Option {
	return func(o option) option {
		o.fsys = fsys
		return o
	}
}

type option struct {
	fsys FS
	join Join
	perm os.FileMode
}

func newOpt(opts ...Option) option {
	o := option{}
	for _, opt := range opts {
		if opt != nil {
			o = opt(o)
		}
	}

	if o.fsys == nil {
		o.fsys = OS()
	}
	if o.join == nil {
		o.join = filepath.Join
	}
	if o.perm == 0 {
		o.perm = RwRR
	}
	return o
}

// CopyFile copies a provided file from src to dest with a default permission of 0o644. It fails if it's a directory.
func CopyFile(src, dest string, opts ...Option) error {
	o := newOpt(opts...)

	// read file from fsys (OperatingFS or specific fsys)
	sfile, err := o.fsys.Open(src)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer sfile.Close()

	// create dest in OS filesystem and not given fsys
	dfile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	defer dfile.Close()

	// copy buffer from src to dest
	if _, err := io.Copy(dfile, sfile); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	// update dest permissions
	if err := dfile.Chmod(o.perm); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}
	return nil
}

// SafeMove moves src into dest while taking care of potentially running process for dest.
func SafeMove(src, dest string, opts ...Option) error {
	o := newOpt(opts...)

	bytes, err := o.fsys.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), RwxRxRxRx); err != nil {
		return fmt.Errorf("mkdir all: %w", err)
	}

	tdest := dest + "_"
	if err := os.WriteFile(tdest, bytes, o.perm); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	if err := os.Rename(tdest, dest); err != nil {
		return fmt.Errorf("move: %w", err)
	}
	return nil
}

// Exists returns a boolean indicating whether the provided input src exists or not.
func Exists(src string, opts ...Option) bool {
	o := newOpt(opts...)

	// read file from fsys (OperatingFS or specific fsys)
	file, err := o.fsys.Open(src)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}
