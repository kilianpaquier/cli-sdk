package upgrade

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/kilianpaquier/cli-sdk/pkg/clog"
)

var (
	_majorRegexp = regexp.MustCompile("^v[0-9]+$")
	_minorRegexp = regexp.MustCompile(`^v[0-9]+\.[0-9]+$`)
	_wordRegexp  = regexp.MustCompile(`[a-zA-Z]+`)
)

var (
	// ErrMajorMinorExclusive is the error returned when both options WithMajor and WithMinor are given and non empty.
	ErrMajorMinorExclusive = errors.New("both major and minor option are mutually exclusive")

	// ErrInvalidOptions is the error returned when there's at least one invalid option.
	ErrInvalidOptions = errors.New("invalid options")
)

// RunOption is the right function to tune Run function with specific behaviors.
type RunOption func(*option) error

// WithAssetTemplate specifies the asset name to match (equal) for during asset finding (to retrieve the appropriate one to install).
//
// By default it's '{{ .Repo }}_{{ .GOOS }}_{{ .GOARCH }}{{ .ArchiveExt }}'.
//
// Various functions are available: 'lower', 'title', 'upper'.
//
// Various variables are available: 'ArchiveExt', 'BinExt', 'GOOS', 'GOARCH', 'Opts' (.Major, .Minor, .Prereleases), 'Repo', 'Tag'.
func WithAssetTemplate(assetTemplate string) RunOption {
	return func(o *option) error {
		o.AssetTemplate = assetTemplate
		return nil
	}
}

// WithDestination defines the output dir where binaries will be downloaded.
//
// By default, installation destination is ${HOME}/.local/bin.
func WithDestination(destdir string) RunOption {
	return func(o *option) error {
		o.Destdir = destdir
		return nil
	}
}

// WithHTTPClient specifies the http client to use for both GetReleases function
// and asset(s) download(s).
//
// By default cleanhttp.DefaultClient() will be used.
func WithHTTPClient(client *http.Client) RunOption {
	return func(o *option) error {
		o.HTTPClient = client
		return nil
	}
}

// WithLogger defines the logger implementation for Run function.
//
// When not provided, no logging will be made.
func WithLogger(log clog.Logger) RunOption {
	return func(o *option) error {
		o.Log = log
		return nil
	}
}

// WithTargetTemplate specifies the target name of the installed binary.
//
// By default it's
// {{- .Repo }}
// {{- if ne .Opts.Major "" }}{{ print "-" .Opts.Major }}
// {{- else if ne .Opts.Minor "" }}{{ print "-" .Opts.Minor }}
// {{- end }}
// {{- if and .Opts.Prereleases (ne .Prerelease "") }}{{ print "-" .Prerelease }}{{ end }}
// {{- .BinExt }}
// which gives 'repo-pre' or 'repo-v1.exe', or 'repo.exe' or 'repo' or 'repo-v1.6', etc.
// depending on inputs options and whether the installed version is a prerelease or not.
//
// Various functions are available: 'lower', 'title', 'upper'.
//
// Various variables are available: 'ArchiveExt', 'BinExt', 'GOOS', 'GOARCH', 'Opts' (with inputs WithMajor, WithMinor and WithPrerelease), 'Repo', 'Tag'.
//
// Note that it's not recommended to use this option since if badly defined a prerelease installation could override the latest stable installation
// or an old installation could override it too, etc.
// Make sure you're aware of unexpected overrides.
func WithTargetTemplate(targetTemplate string) RunOption {
	return func(o *option) error {
		o.TargetTemplate = targetTemplate
		return nil
	}
}

// WithMajor specifies if upgraded / installed package must concern a specific major version.
//
// By default all major versions can be used (outside of prereleases which can be included with WithPrerelease).
func WithMajor(major string) RunOption {
	return func(o *option) error {
		o.Major = major
		if o.Major != "" && !_majorRegexp.MatchString(major) {
			return fmt.Errorf("invalid major version '%s'", major)
		}
		return nil
	}
}

// WithMinor specifies if upgraded / installed package must concern a specific minor version.
//
// By default all minor versions can be used (outside of prereleases which can be included with WithPrerelease).
func WithMinor(minor string) RunOption {
	return func(o *option) error {
		o.Minor = minor
		if o.Minor != "" && !_minorRegexp.MatchString(minor) {
			return fmt.Errorf("invalid minor version '%s'", minor)
		}
		return nil
	}
}

// WithPrereleases specifies whether prerelease versions can be considered for upgrade / installation.
func WithPrereleases(accepted bool) RunOption {
	return func(o *option) error {
		o.Prereleases = accepted
		return nil
	}
}

// option is the struct related to Option function(s) defining all optional properties.
type option struct {
	releaseOptions

	AssetTemplate  string
	Destdir        string
	HTTPClient     *http.Client
	Log            clog.Logger
	TargetTemplate string
}

// newOpt creates a new option struct with all input Option functions
// while taking care of default values.
//
// It returns an error in case some input options are invalid.
func newOpt(opts ...RunOption) (option, error) {
	o := &option{}

	var errs []error
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(o); err != nil {
			errs = append(errs, err)
		}
	}
	// ensure major and minor aren't given together since there're mutually exclusive
	if o.Major != "" && o.Minor != "" {
		errs = append(errs, ErrMajorMinorExclusive)
	}
	if len(errs) > 0 {
		errs = append(errs, ErrInvalidOptions)
	}

	if o.AssetTemplate == "" {
		o.AssetTemplate = `{{ .Repo }}_{{ .GOOS }}_{{ .GOARCH }}{{ .ArchiveExt }}`
	}
	if o.Destdir == "" {
		home, _ := os.UserHomeDir()
		o.Destdir = filepath.Join(home, ".local", "bin")
	}
	if o.HTTPClient == nil {
		o.HTTPClient = cleanhttp.DefaultClient()
	}
	if o.Log == nil {
		o.Log = clog.Noop()
	}
	if o.TargetTemplate == "" {
		o.TargetTemplate = `
{{- .Repo }}
{{- if ne .Opts.Major "" }}{{ print "-" .Opts.Major }}
{{- else if ne .Opts.Minor "" }}{{ print "-" .Opts.Minor }}
{{- end }}

{{- if and .Opts.Prereleases (ne .Prerelease "") }}{{ print "-" .Prerelease }}{{ end }}
{{- .BinExt }}`
	}

	return *o, errors.Join(errs...)
}
