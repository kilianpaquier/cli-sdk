package upgrade

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/samber/lo"
	"golang.org/x/mod/semver"

	cfs "github.com/kilianpaquier/cli-sdk/pkg/fs"
)

var (
	// ErrNoGetRelease is the error returned by Run when the input getReleases isn't given.
	ErrNoGetRelease = errors.New("getReleases func must not be nil")

	// ErrNoProjectName is the erreur returned by Run when the input projectName isn't given.
	ErrNoProjectName = errors.New("projectName must not be empty")
)

// Release represents a release with its assets, its name
// and other useful properties.
type Release struct {
	Assets  []Asset
	TagName string
}

// Asset represents a release asset with its download URL and its name.
type Asset struct {
	DownloadURL string
	Name        string
}

// GetReleases is the signature function to give to WithGetReleases.
//
// It can be useful in case of a package not hosted by github (hence GithubReleases would not be appropriate)
// to avoid redeveloping the whole feature.
type GetReleases func(ctx context.Context, httpClient *http.Client) ([]Release, error)

// Run is the main function of upgrade package.
// It reads all releases from the provided GetReleases function,
// searches the appropriate release to install depending on input filters (major, minor, prerelease)
// and then installs it if found (or else does nothing).
//
// Installation, as provided in various functions docs, is made either in ${HOME}/.local/bin
// or in provided destination directory with WithDestination option.
//
// The final binary will always be of the form {{projectName}}{{version}} taking care
// of adding the right extension depending on target installation platform.
func Run(ctx context.Context, projectName, currentVersion string, getReleases GetReleases, opts ...RunOption) error {
	if projectName == "" {
		return ErrNoProjectName
	}
	if getReleases == nil {
		return ErrNoGetRelease
	}

	// group things related to OS specificities for a better maintenability
	ext := lo.Ternary(runtime.GOOS == "windows", ".exe", "")
	suffix := fmt.Sprint(runtime.GOOS, "_", runtime.GOARCH, lo.Ternary(runtime.GOOS == "windows", ".zip", ".tar.gz")) // linux_amd64.tar.gz or windows_arm64.zip, etc.

	o, err := newOpt(opts...)
	if err != nil {
		return fmt.Errorf("parsing options: %w", err)
	}

	releases, err := getReleases(ctx, o.httpClient)
	if err != nil {
		return fmt.Errorf("get releases: %w", err)
	}

	release, ok := findRelease(releases, o.releaseOptions)
	if !ok {
		o.log.Info("no new version found matching options")
		return nil
	}
	name := binaryName(projectName, ext, release.TagName, o.releaseOptions)
	dest := filepath.Join(o.destdir, name)

	o.log.Infof("installing version '%s'", release.TagName)
	if currentVersion == release.TagName && cfs.Exists(dest) {
		o.log.Infof("version '%s' already installed in '%s'", release.TagName, dest)
		return nil
	}

	url, err := getDownloadURL(release, suffix)
	if err != nil {
		return fmt.Errorf("get download url: %w", err)
	}

	httpGetter := getter.HttpGetter{
		Client:                o.httpClient,
		XTerraformGetDisabled: true, // unnecessary for assets downloading case
	}
	options := []getter.ClientOption{
		getter.WithContext(ctx),
		getter.WithGetters(map[string]getter.Getter{"http": &httpGetter, "https": &httpGetter}),
		func(c *getter.Client) error { c.DisableSymlinks = true; return nil }, // security consideration
	}
	// download in temporary directory the release (since we only want to move, rename and keep the binary)
	tmp := filepath.Join(os.TempDir(), projectName, release.TagName)
	if err := getter.Get(tmp, url, options...); err != nil {
		return fmt.Errorf("download asset(s): %w", err)
	}

	// move safely (as the current binary could be running) the newest version in place
	if err := cfs.SafeMove(filepath.Join(tmp, projectName+ext), dest, cfs.WithPerm(cfs.RwxRxRxRx)); err != nil {
		return fmt.Errorf("safe move: %w", err)
	}
	o.log.Infof("successfully installed version '%s' in '%s'", release.TagName, dest)
	return nil
}

// releaseOptions is the struct will all options for releases filtering.
type releaseOptions struct {
	major      string
	minor      string
	prerelease bool
}

// findRelease finds the appropriate release to install in the input slice of releases depending on search version and provided options.
func findRelease(releases []Release, opts releaseOptions) (*Release, bool) {
	// remove all invalid semver releases or draft releases
	candidates := slices.DeleteFunc(releases, func(r Release) bool { return !semver.IsValid(r.TagName) })

	// keep only versions related to given major version
	if opts.major != "" {
		candidates = slices.DeleteFunc(candidates, func(r Release) bool { return semver.Major(r.TagName) != opts.major })
	}
	// keep only versions related to given minor version
	if opts.minor != "" {
		candidates = slices.DeleteFunc(candidates, func(r Release) bool { return semver.MajorMinor(r.TagName) != opts.minor })
	}

	// sort all appropriate releases
	slices.SortStableFunc(candidates, func(r1, r2 Release) int {
		return semver.Compare(r1.TagName, r2.TagName)
	})

	// loop over the slice in reverse mode to retrieve the first appropriate version
	// depending on whether prereleases are accepted or not
	for i := len(candidates) - 1; i >= 0; i-- {
		found := candidates[i]

		// if prereleases aren't accepted and version is a prerelease
		// continue since it cannot be installed
		if !opts.prerelease && !_completeRegexp.MatchString(found.TagName) {
			continue
		}
		return &found, true
	}
	return nil, false
}

// getDownloadURL returns the right URL to use for downloading a release.
//
// It's a specific function since download is handled by go-getter and that it can handle checksums verification.
// As such, returned URL can be enriched with the checksums URL.
func getDownloadURL(release *Release, suffix string) (string, error) {
	// find the right asset
	asset, ok := lo.Find(release.Assets, func(asset Asset) bool { return strings.HasSuffix(asset.Name, suffix) })
	if !ok {
		return "", fmt.Errorf("no valid release asset found with suffix '%s'", suffix)
	}

	// find checksum file in assets for verification during download
	checksum, ok := lo.Find(release.Assets, func(asset Asset) bool { return asset.Name == "checksums.txt" })
	if ok {
		return fmt.Sprintf("%s?checksum=file:%s", asset.DownloadURL, checksum.DownloadURL), nil
	}
	return asset.DownloadURL, nil
}

// binaryName returns the appropriate name for the binary depending on given release options.
func binaryName(projectName, ext, version string, opts releaseOptions) string {
	sep := "-"

	name := projectName
	if opts.major != "" {
		name += sep + opts.major
	} else if opts.minor != "" {
		name += sep + opts.minor
	}
	if opts.prerelease && !_completeRegexp.MatchString(version) {
		name += sep + "pre"
	}
	return name + ext
}
