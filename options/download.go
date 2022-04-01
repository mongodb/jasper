package options

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/evergreen-ci/bond"
	"github.com/evergreen-ci/utility"
	"github.com/mholt/archiver/v3"
	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// Download represents the options to download a file to a given path and
// optionally extract its contents.
type Download struct {
	URL         string  `json:"url"`
	Path        string  `json:"path"`
	ArchiveOpts Archive `json:"archive_opts"`
}

// Validate checks the download options.
func (opts Download) Validate() error {
	catcher := grip.NewBasicCatcher()

	if opts.URL == "" {
		catcher.New("download URL cannot be empty")
	}

	catcher.ErrorfWhen(!filepath.IsAbs(opts.Path), "download path '%s' must be an absolute path", opts.Path)

	catcher.Wrap(opts.ArchiveOpts.Validate(), "invalid archive options")

	return catcher.Resolve()
}

// Download executes the download operation.
func (opts Download) Download() error {
	req, err := http.NewRequest(http.MethodGet, opts.URL, nil)
	if err != nil {
		return errors.Wrap(err, "building request")
	}

	client := utility.GetHTTPClient()
	defer utility.PutHTTPClient(client)

	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "downloading file for URL '%s'", opts.URL)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("%s: could not download URL '%s' to path '%s'", resp.Status, opts.URL, opts.Path)
	}

	if err = writeFile(resp.Body, opts.Path); err != nil {
		return err
	}

	if opts.ArchiveOpts.ShouldExtract {
		if err = opts.Extract(); err != nil {
			return errors.Wrapf(err, "extracting file '%s' to path '%s'", opts.Path, opts.ArchiveOpts.TargetPath)
		}
	}

	return nil
}

// Extract extracts the download to the path specified, using the archive format
// specified.
func (opts Download) Extract() error {
	var archiveHandler archiver.Unarchiver
	switch opts.ArchiveOpts.Format {
	case ArchiveAuto:
		return archiver.Unarchive(opts.Path, opts.ArchiveOpts.TargetPath)
	case ArchiveTarGz:
		archiveHandler = archiver.DefaultTarGz
	case ArchiveZip:
		archiveHandler = archiver.DefaultZip
	default:
		return errors.Errorf("unrecognized archive format '%s'", opts.ArchiveOpts.Format)
	}

	if err := archiveHandler.Unarchive(opts.Path, opts.ArchiveOpts.TargetPath); err != nil {
		return errors.Wrapf(err, "extracting archive %s to %s", opts.Path, opts.ArchiveOpts.TargetPath)
	}

	return nil
}

// MongoDBDownload represents the options to download MongoDB on a target
// platform.
type MongoDBDownload struct {
	BuildOpts bond.BuildOptions `json:"build_opts"`
	Path      string            `json:"path"`
	Releases  []string          `json:"releases"`
}

// Validate checks for valid MongoDB download options.
func (opts MongoDBDownload) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.ErrorfWhen(!filepath.IsAbs(opts.Path), "download path '%s' must be an absolute path", opts.Path)
	catcher.Wrap(opts.BuildOpts.Validate(), "invalid MongoDB build options")

	return catcher.Resolve()
}

// Cache represent the configuration options for the LRU cache.
type Cache struct {
	Disabled   bool          `json:"disabled"`
	PruneDelay time.Duration `json:"prune_delay"`
	MaxSize    int           `json:"max_size"`
}

// Validate checks for valid cache options.
func (opts Cache) Validate() error {
	catcher := grip.NewBasicCatcher()

	catcher.NewWhen(opts.MaxSize < 0, "max size cannot be negative")
	catcher.NewWhen(opts.PruneDelay < 0, "prune delay cannot be negative")

	return catcher.Resolve()
}
