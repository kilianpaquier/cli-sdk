package upgrade

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
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
type RunOption func(*runOptions) error

// WithAssetTemplate specifies the asset name to match (equal) for during asset finding (to retrieve the appropriate one to install).
//
// By default it's:
//
//	{{ .Repo }}_{{ .GOOS }}_{{ .GOARCH }}{{ .ArchiveExt }}
//
// Various functions are available: 'lower', 'title', 'upper'.
//
// Various variables are available: 'ArchiveExt', 'BinExt', 'GOOS', 'GOARCH', 'Opts' (.Major, .Minor, .Prereleases), 'Repo', 'Tag'.
func WithAssetTemplate(assetTemplate string) RunOption {
	return func(o *runOptions) error {
		o.assetTemplate = assetTemplate
		return nil
	}
}

// WithDestination defines the output dir where binaries will be downloaded.
//
// By default, installation destination is ${HOME}/.local/bin.
func WithDestination(destdir string) RunOption {
	return func(o *runOptions) error {
		o.destdir = destdir
		return nil
	}
}

// WithHTTPClient specifies the http client to use for both GetReleases function
// and asset(s) download(s).
//
// By default cleanhttp.DefaultClient() will be used.
func WithHTTPClient(client *http.Client) RunOption {
	return func(o *runOptions) error {
		o.httpClient = client
		return nil
	}
}

// WithTargetTemplate specifies the target name of the installed binary.
//
// By default it's
//
//	{{- .Repo }}
//	{{- if ne .Opts.Major "" }}{{ print "-" .Opts.Major }}
//	{{- else if ne .Opts.Minor "" }}{{ print "-" .Opts.Minor }}
//	{{- end }}
//	{{- if and .Opts.Prereleases (ne .Prerelease "") }}{{ print "-" .Prerelease }}{{ end }}
//	{{- .BinExt }}
//
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
	return func(ro *runOptions) error {
		ro.targetTemplate = targetTemplate
		return nil
	}
}

// WithMajor specifies if upgraded / installed package must concern a specific major version.
//
// By default all major versions can be used (outside of prereleases which can be included with WithPrerelease).
func WithMajor(major string) RunOption {
	return func(ro *runOptions) error {
		ro.Major = major
		if ro.Major != "" && !_majorRegexp.MatchString(major) {
			return fmt.Errorf("invalid major version '%s'", major)
		}
		return nil
	}
}

// WithMinor specifies if upgraded / installed package must concern a specific minor version.
//
// By default all minor versions can be used (outside of prereleases which can be included with WithPrerelease).
func WithMinor(minor string) RunOption {
	return func(ro *runOptions) error {
		ro.Minor = minor
		if ro.Minor != "" && !_minorRegexp.MatchString(minor) {
			return fmt.Errorf("invalid minor version '%s'", minor)
		}
		return nil
	}
}

// WithPrereleases specifies whether prerelease versions can be considered for upgrade / installation.
func WithPrereleases(accepted bool) RunOption {
	return func(ro *runOptions) error {
		ro.Prereleases = accepted
		return nil
	}
}

// runOptions is the struct related to Option function(s) defining all optional properties.
type runOptions struct {
	releaseOptions

	assetTemplate  string
	destdir        string
	httpClient     *http.Client
	targetTemplate string
}

// newRunOpt creates a new option struct with all input Option functions
// while taking care of default values.
//
// It returns an error in case some input options are invalid.
func newRunOpt(opts ...RunOption) (runOptions, error) {
	var errs []error
	var ro runOptions
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(&ro); err != nil {
			errs = append(errs, err)
		}
	}
	// ensure major and minor aren't given together since there're mutually exclusive
	if ro.Major != "" && ro.Minor != "" {
		errs = append(errs, ErrMajorMinorExclusive)
	}
	if len(errs) > 0 {
		errs = slices.Insert(errs, 0, ErrInvalidOptions)
		ef := make([]any, 0, len(errs))
		wraps := make([]string, 0, len(errs)) // it's uggly but errors.Join with Error prints with '\n' and it's not customizable
		for _, err := range errs {
			ef = append(ef, err)
			wraps = append(wraps, "%w")
		}
		return ro, fmt.Errorf(strings.Join(wraps, ": "), ef...)
	}

	if ro.assetTemplate == "" {
		ro.assetTemplate = `{{ .Repo }}_{{ .GOOS }}_{{ .GOARCH }}{{ .ArchiveExt }}`
	}
	if ro.destdir == "" {
		home, _ := os.UserHomeDir()
		ro.destdir = filepath.Join(home, ".local", "bin")
	}
	if ro.httpClient == nil {
		ro.httpClient = cleanhttp.DefaultClient()
	}
	if ro.targetTemplate == "" {
		ro.targetTemplate = `
{{- .Repo }}
{{- if ne .Opts.Major "" }}{{ print "-" .Opts.Major }}
{{- else if ne .Opts.Minor "" }}{{ print "-" .Opts.Minor }}
{{- end }}
{{- if and .Opts.Prereleases (ne .Prerelease "") }}{{ print "-" .Prerelease }}{{ end }}
{{- .BinExt }}`
	}

	return ro, nil
}
