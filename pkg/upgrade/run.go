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

	getter "github.com/hashicorp/go-getter/v2"
	"golang.org/x/mod/semver"

	"github.com/kilianpaquier/cli-sdk/pkg/cfs"
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
func Run(ctx context.Context, repo, currentVersion string, getReleases GetReleases, opts ...RunOption) error {
	if repo == "" {
		return ErrNoProjectName
	}
	if getReleases == nil {
		return ErrNoGetRelease
	}

	o, err := newOpt(opts...)
	if err != nil {
		return fmt.Errorf("parsing options: %w", err)
	}

	releases, err := getReleases(ctx, o.HTTPClient)
	if err != nil {
		return fmt.Errorf("get releases: %w", err)
	}

	release, ok := findRelease(releases, o.releaseOptions)
	if !ok {
		o.Log.Infof("no new version found matching options")
		return nil
	}

	s := _wordRegexp.FindAllString(semver.Prerelease(release.TagName), -1)
	var prerelease string
	if len(s) > 0 {
		// retrieve only the first element, in case there's '-beta.toto', etc. (weird cases)
		// in any case semver.Prerelease already does the job to retrieve '-beta' with for instance v1.5.6-beta+meta
		// but semver.Prerelease was missing the case of retrieving '-beta' with v1.5.6-beta.1, where it returned '-beta.1'
		prerelease = s[0]
	}
	templateData := map[string]any{
		"ArchiveExt": archiveExt(),
		"BinExt":     binExt(),
		"GOARCH":     runtime.GOARCH,
		"GOOS":       runtime.GOOS,
		"Opts":       o.releaseOptions,
		"Prerelease": prerelease,
		"Repo":       repo,
		"Tag":        release.TagName,
	}

	targetName, err := getTemplateValue(o.TargetTemplate, templateData)
	if err != nil {
		return fmt.Errorf("get target name: %w", err)
	}
	dest := filepath.Join(o.Destdir, targetName)

	o.Log.Infof("installing version '%s'", release.TagName)
	if currentVersion == release.TagName && cfs.Exists(dest) {
		o.Log.Infof("version '%s' already installed in '%s'", release.TagName, dest)
		return nil
	}

	assetName, err := getTemplateValue(o.AssetTemplate, templateData)
	if err != nil {
		return fmt.Errorf("get asset name: %w", err)
	}

	url, err := getDownloadURL(release, assetName)
	if err != nil {
		return fmt.Errorf("get download url: %w", err)
	}

	get := getter.Client{
		DisableSymlinks: true,
		Getters:         []getter.Getter{&getter.HttpGetter{Client: o.HTTPClient, XTerraformGetDisabled: true}},
	}
	// download in temporary directory the release (since we only want to move, rename and keep the binary)
	tmp := filepath.Join(os.TempDir(), repo, release.TagName)
	if _, err := get.Get(ctx, &getter.Request{Src: url, Dst: tmp, GetMode: getter.ModeAny}); err != nil {
		return fmt.Errorf("download asset(s): %w", err)
	}

	// move safely (as the current binary could be running) the newest version in place
	if err := cfs.SafeMove(filepath.Join(tmp, repo+binExt()), dest, cfs.WithPerm(cfs.RwxRxRxRx)); err != nil {
		return fmt.Errorf("safe move: %w", err)
	}
	o.Log.Infof("successfully installed version '%s' in '%s'", release.TagName, dest)
	return nil
}

// releaseOptions is the struct will all options for releases filtering.
type releaseOptions struct {
	Prereleases bool
	Major       string
	Minor       string
}

// findRelease finds the appropriate release to install in the input slice of releases depending on search version and provided options.
func findRelease(releases []Release, opts releaseOptions) (*Release, bool) {
	// remove all invalid semver releases or draft releases
	candidates := slices.DeleteFunc(releases, func(r Release) bool { return !semver.IsValid(r.TagName) })

	// keep only versions related to given major version
	if opts.Major != "" {
		candidates = slices.DeleteFunc(candidates, func(r Release) bool { return semver.Major(r.TagName) != opts.Major })
	}
	// keep only versions related to given minor version
	if opts.Minor != "" {
		candidates = slices.DeleteFunc(candidates, func(r Release) bool { return semver.MajorMinor(r.TagName) != opts.Minor })
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
		if !opts.Prereleases && semver.Prerelease(found.TagName) != "" {
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
func getDownloadURL(release *Release, assetName string) (string, error) {
	var bin, checksum Asset
	for _, asset := range release.Assets {
		// find the right asset
		if asset.Name == assetName {
			bin = asset
		}

		// find checksum file in assets for verification during download
		if asset.Name == "checksums.txt" {
			checksum = asset
		}
	}

	if bin == (Asset{}) {
		return "", fmt.Errorf("no valid release asset found with suffix '%s'", assetName)
	}

	if checksum != (Asset{}) {
		return fmt.Sprintf("%s?checksum=file:%s", bin.DownloadURL, checksum.DownloadURL), nil
	}
	return bin.DownloadURL, nil
}

// binExt returns the appropriate extension for a binary depending on current GOOS.
func binExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// archiveExt returns the appropriate extension for an archive depending on current GOOS.
func archiveExt() string {
	if runtime.GOOS == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}
