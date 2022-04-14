package options

import (
	"path/filepath"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// ArchiveFormat represents an archive file type.
type ArchiveFormat string

const (
	// ArchiveAuto is an ArchiveFormat that does not force any particular type of
	// archive format.
	ArchiveAuto ArchiveFormat = "auto"
	// ArchiveTarGz is an ArchiveFormat for gzipped tar archives.
	ArchiveTarGz ArchiveFormat = "targz"
	// ArchiveZip is an ArchiveFormat for Zip archives.
	ArchiveZip ArchiveFormat = "zip"
)

// Validate checks that the ArchiveFormat is a recognized format.
func (f ArchiveFormat) Validate() error {
	switch f {
	case ArchiveTarGz, ArchiveZip, ArchiveAuto:
		return nil
	default:
		return errors.Errorf("unknown archive format %s", f)
	}
}

// Archive encapsulates options related to management of archive files.
type Archive struct {
	ShouldExtract bool
	Format        ArchiveFormat
	TargetPath    string
}

// Validate checks the archive file options.
func (opts Archive) Validate() error {
	if !opts.ShouldExtract {
		return nil
	}

	catcher := grip.NewBasicCatcher()

	catcher.ErrorfWhen(!filepath.IsAbs(opts.TargetPath), "download path '%s' must be an absolute path", opts.TargetPath)
	catcher.Wrap(opts.Format.Validate(), "invalid archive format")

	return catcher.Resolve()
}
