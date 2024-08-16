package upgrade

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/go-cleanhttp"

	"github.com/kilianpaquier/cli-sdk/pkg/logger"
)

var (
	_majorRegexp    = regexp.MustCompile("^v[0-9]+$")
	_minorRegexp    = regexp.MustCompile(`^v[0-9]+\.[0-9]+$`)
	_completeRegexp = regexp.MustCompile(`^v[0-9]+(\.[0-9]+){2}$`)
)

// ErrMajorMinorExclusive is the error returned when both options WithMajor and WithMinor are given and non empty.
var ErrMajorMinorExclusive = errors.New("both major and minor option are mutually exclusive")

// RunOption is the right function to tune Run function with specific behaviors.
type RunOption func(*option) error

// WithDestination defines the output dir where binaries will be downloaded.
//
// By default, when not given, ${HOME}/.local/bin/{{repo}}{{version}} will be used (extension is preserved).
// Note that {{repo}} is given as input in Run function.
func WithDestination(destdir string) RunOption {
	return func(o *option) error {
		if destdir == "" {
			return nil
		}
		o.destdir = destdir
		return nil
	}
}

// WithHTTPClient specifies the http client to use for both GetReleases function
// and asset(s) download(s).
//
// By default cleanhttp.DefaultClient() will be used.
func WithHTTPClient(client *http.Client) RunOption {
	return func(o *option) error {
		o.httpClient = client
		return nil
	}
}

// WithLogger defines the logger implementation for Run function.
//
// When not provided, the default one used is the one from std log library.
func WithLogger(log logger.Logger) RunOption {
	return func(o *option) error {
		o.log = log
		return nil
	}
}

// WithMajor specifies if upgraded / installed package must concern a specific major version.
//
// By default all major versions can be used (outside of prereleases which can be included with WithPrerelease).
func WithMajor(major string) RunOption {
	return func(o *option) error {
		if major == "" {
			return nil
		}

		if !_majorRegexp.MatchString(major) {
			return fmt.Errorf("invalid major version '%s'", major)
		}

		o.major = major
		return nil
	}
}

// WithMinor specifies if upgraded / installed package must concern a specific minor version.
//
// By default all minor versions can be used (outside of prereleases which can be included with WithPrerelease).
func WithMinor(minor string) RunOption {
	return func(o *option) error {
		if minor == "" {
			return nil
		}
		if !_minorRegexp.MatchString(minor) {
			return fmt.Errorf("invalid minor version '%s'", minor)
		}

		o.minor = minor
		return nil
	}
}

// WithPrerelease specifies whether prerelease versions can be considered for upgrade / installation.
func WithPrerelease(prerelease bool) RunOption {
	return func(o *option) error {
		o.prerelease = prerelease
		return nil
	}
}

// option is the struct related to Option function(s) defining all optional properties.
type option struct {
	releaseOptions

	destdir    string
	httpClient *http.Client
	log        logger.Logger
}

// newOpt creates a new option struct with all input Option functions
// while taking care of default values.
//
// It returns an error in case some input options are invalid.
func newOpt(opts ...RunOption) (option, error) {
	o := &option{}

	var errs []error
	for _, opt := range opts {
		if opt != nil {
			errs = append(errs, opt(o))
		}
	}

	// ensure major and minor aren't given together since there're mutually exclusive
	if o.major != "" && o.minor != "" {
		errs = append(errs, ErrMajorMinorExclusive)
	}

	if o.destdir == "" {
		home, _ := os.UserHomeDir()
		o.destdir = filepath.Join(home, ".local", "bin")
	}
	if o.httpClient == nil {
		o.httpClient = cleanhttp.DefaultClient()
	}
	if o.log == nil {
		o.log = logger.Std()
	}
	return *o, errors.Join(errs...)
}
