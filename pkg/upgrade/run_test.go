package upgrade_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-github/v63/github"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kilianpaquier/cli-sdk/pkg/cfs"
	"github.com/kilianpaquier/cli-sdk/pkg/clog"
	"github.com/kilianpaquier/cli-sdk/pkg/upgrade"
)

func toPtr[T any](in T) *T {
	return &in
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	log := clog.Std()

	// setup github / go-getter mocking
	httpClient := cleanhttp.DefaultClient()
	httpmock.ActivateNonDefault(httpClient)
	t.Cleanup(httpmock.DeactivateAndReset)

	getReleases := upgrade.GithubReleases("owner", "repo")

	t.Run("error_missing_project_name", func(t *testing.T) {
		// Act
		err := upgrade.Run(ctx, "", "", nil)

		// Assert
		assert.ErrorIs(t, err, upgrade.ErrNoProjectName)
	})

	t.Run("error_missing_get_releases", func(t *testing.T) {
		// Act
		err := upgrade.Run(ctx, "repo", "", nil)

		// Assert
		assert.ErrorIs(t, err, upgrade.ErrNoGetRelease)
	})

	t.Run("error_invalid_major_option", func(t *testing.T) {
		// Act
		err := upgrade.Run(ctx, "repo", "", getReleases, upgrade.WithMajor("invalid"))

		// Assert
		assert.ErrorContains(t, err, "parsing options")
		assert.ErrorContains(t, err, "invalid major version")
	})

	t.Run("error_invalid_minor_option", func(t *testing.T) {
		// Act
		err := upgrade.Run(ctx, "repo", "", getReleases, upgrade.WithMinor("invalid"))

		// Assert
		assert.ErrorContains(t, err, "parsing options")
		assert.ErrorContains(t, err, "invalid minor version")
	})

	t.Run("error_both_major_minor_options_given", func(t *testing.T) {
		// Act
		err := upgrade.Run(ctx, "repo", "", getReleases, upgrade.WithMajor("v1"), upgrade.WithMinor("v4.3"))

		// Assert
		assert.ErrorContains(t, err, "parsing options")
		assert.ErrorContains(t, err, upgrade.ErrMajorMinorExclusive.Error())
	})

	t.Run("error_get_releases_custom", func(t *testing.T) {
		// Arrange
		var errReleases upgrade.GetReleases = func(_ context.Context, _ *http.Client) ([]upgrade.Release, error) {
			return nil, errors.New("some error")
		} // specify var with type to ensure interface is implemented

		// Act
		err := upgrade.Run(ctx, "repo", "", errReleases)

		// Assert
		assert.ErrorContains(t, err, "get releases: some error")
	})

	t.Run("error_get_releases_github", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		url := "https://api.github.com/repos/owner/repo/releases?page=1&per_page=100"
		httpmock.RegisterResponder(http.MethodGet, url,
			httpmock.NewStringResponder(http.StatusInternalServerError, "unused error"))

		// Act
		err := upgrade.Run(ctx, "repo", "", getReleases, upgrade.WithHTTPClient(httpClient))

		// Assert
		assert.ErrorContains(t, err, fmt.Sprintf("get releases: list releases: %s %s", http.MethodGet, url))
	})

	t.Run("error_no_valid_asset", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		url := "https://api.github.com/repos/owner/repo/releases?page=1&per_page=100"
		httpmock.RegisterResponder(http.MethodGet, url,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{
				{
					TagName: toPtr("v1.0.0"),
					Assets: []*github.ReleaseAsset{
						{Name: toPtr("suffix missing"), BrowserDownloadURL: toPtr("some URL")},
					},
				},
			}))

		// Act
		err := upgrade.Run(ctx, "repo", "", getReleases, upgrade.WithHTTPClient(httpClient))

		// Assert
		assert.ErrorContains(t, err, "get download url")
		assert.ErrorContains(t, err, "no valid release asset found")
	})

	t.Run("error_download_assets", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		url := "https://api.github.com/repos/owner/repo/releases?page=1&per_page=100"
		httpmock.RegisterResponder(http.MethodGet, url,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{
				{
					TagName: toPtr("v1.0.0"),
					Assets: []*github.ReleaseAsset{
						{Name: toPtr(fmt.Sprintf("%s_%s.zip", runtime.GOOS, runtime.GOARCH)), BrowserDownloadURL: toPtr("http://example.com/asset/download")},
						{Name: toPtr(fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)), BrowserDownloadURL: toPtr("http://example.com/asset/download")},
					},
				},
			}))

		// Act
		err := upgrade.Run(ctx, "repo", "", getReleases, upgrade.WithHTTPClient(httpClient))

		// Assert
		assert.ErrorContains(t, err, "download asset(s)")
		assert.ErrorContains(t, err, "http://example.com/asset/download")
	})

	t.Run("error_safe_move", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		releasesURL := "https://api.github.com/repos/owner/repo/releases?page=1&per_page=100"
		downloadURL := "http://example.com/asset/download/repo"
		httpmock.RegisterResponder(http.MethodGet, releasesURL,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{
				{
					TagName: toPtr("v1.0.0"),
					Assets: []*github.ReleaseAsset{
						{Name: toPtr(fmt.Sprintf("%s_%s.zip", runtime.GOOS, runtime.GOARCH)), BrowserDownloadURL: &downloadURL},
						{Name: toPtr(fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)), BrowserDownloadURL: &downloadURL},
					},
				},
			}))
		httpmock.RegisterResponder(http.MethodGet, downloadURL,
			httpmock.NewStringResponder(http.StatusOK, "some text for a file")) // go-getter doesn't really download anything sadly

		// Act
		err := upgrade.Run(ctx, "repo", "v0.0.0", getReleases,
			upgrade.WithDestination(t.TempDir()),
			upgrade.WithHTTPClient(httpClient),
			upgrade.WithLogger(log))

		// Assert
		assert.ErrorContains(t, err, "safe move")
	})

	t.Run("success_no_appropriate_release", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		url := "https://api.github.com/repos/owner/repo/releases?page=1&per_page=100"
		httpmock.RegisterResponder(http.MethodGet, url,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{
				{TagName: toPtr("v1.0.0-beta.1")},
			}))

		dest := filepath.Join(t.TempDir(), "subdir")

		// Act
		err := upgrade.Run(ctx, "repo", "", getReleases,
			upgrade.WithDestination(dest),
			upgrade.WithHTTPClient(httpClient),
			upgrade.WithLogger(log))

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, httpmock.GetTotalCallCount(), 1)
		assert.NoDirExists(t, dest)
	})

	t.Run("success_already_installed", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		releasesURL := "https://api.github.com/repos/owner/repo/releases?page=1&per_page=100"
		downloadURL := "http://example.com/asset/download/repo"
		httpmock.RegisterResponder(http.MethodGet, releasesURL,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{
				{
					TagName: toPtr("v1.0.1-beta.1"),
					Assets: []*github.ReleaseAsset{
						{Name: toPtr(fmt.Sprintf("%s_%s.zip", runtime.GOOS, runtime.GOARCH)), BrowserDownloadURL: &downloadURL},
						{Name: toPtr(fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)), BrowserDownloadURL: &downloadURL},
					},
				},
			}))
		httpmock.RegisterResponder(http.MethodGet, downloadURL,
			httpmock.NewStringResponder(http.StatusOK, "some text for a file")) // go-getter doesn't really download anything sadly

		dest := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dest, "repo-pre"), cfs.RwxRxRxRx))
		require.NoError(t, os.MkdirAll(filepath.Join(dest, "repo-pre.exe"), cfs.RwxRxRxRx))

		// Act
		err := upgrade.Run(ctx, "repo", "v1.0.1-beta.1", getReleases,
			upgrade.WithDestination(dest),
			upgrade.WithHTTPClient(httpClient),
			upgrade.WithLogger(log),
			upgrade.WithPrerelease(true))

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, httpmock.GetTotalCallCount(), 1)
	})

	t.Run("success_download", func(t *testing.T) {
		// Arrange
		t.Cleanup(httpmock.Reset)
		releasesURL := "https://api.github.com/repos/owner/repo/releases?page=1&per_page=100"
		downloadURL := "http://example.com/asset/download/repo"
		httpmock.RegisterResponder(http.MethodGet, releasesURL,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{
				{
					TagName: toPtr("v1.0.0"),
					Assets: []*github.ReleaseAsset{
						{Name: toPtr(fmt.Sprintf("%s_%s.zip", runtime.GOOS, runtime.GOARCH)), BrowserDownloadURL: &downloadURL},
						{Name: toPtr(fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)), BrowserDownloadURL: &downloadURL},
					},
				},
			}))
		httpmock.RegisterResponder(http.MethodGet, downloadURL,
			httpmock.NewStringResponder(http.StatusOK, "some text for a file")) // go-getter doesn't really download anything sadly

		tmp := filepath.Join(os.TempDir(), "repo", "v1.0.0")
		require.NoError(t, os.MkdirAll(tmp, cfs.RwxRxRxRx))
		t.Cleanup(func() { assert.NoError(t, os.RemoveAll(filepath.Dir(tmp))) })
		if runtime.GOOS == "windows" {
			file, err := os.Create(filepath.Join(tmp, "repo.exe")) // create windows file
			require.NoError(t, err)
			require.NoError(t, file.Close())
		} else {
			file, err := os.Create(filepath.Join(tmp, "repo")) // create linux file
			require.NoError(t, err)
			require.NoError(t, file.Close())
		}

		// Act
		err := upgrade.Run(ctx, "repo", "v0.0.0", getReleases,
			upgrade.WithDestination(t.TempDir()),
			upgrade.WithHTTPClient(httpClient),
			upgrade.WithLogger(log))

		// Assert
		assert.NoError(t, err)
	})
}

func TestFindRelease(t *testing.T) {
	t.Run("success_no_releases", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{}
		options := upgrade.NewReleaseOptions("", "", false)

		// Act
		_, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.False(t, ok)
	})

	t.Run("success_invalid_semver", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{
			{TagName: "no_semver"},
		}
		options := upgrade.NewReleaseOptions("", "", false)

		// Act
		_, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.False(t, ok)
	})

	t.Run("success_major_option_older", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{
			{TagName: "v1.7.8"},
			{TagName: "v2.0.0"},
			{TagName: "v2.0.5"},
		}
		options := upgrade.NewReleaseOptions("v2", "", false)

		// Act
		release, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, &upgrade.Release{TagName: "v2.0.5"}, release)
	})

	t.Run("success_major_option_newer", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{
			{TagName: "v2.7.8"},
			{TagName: "v3.0.0"},
			{TagName: "v3.0.5"},
		}
		options := upgrade.NewReleaseOptions("v3", "", false)

		// Act
		release, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, &upgrade.Release{TagName: "v3.0.5"}, release)
	})

	t.Run("success_minor_option_older", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{
			{TagName: "v2.3.8"},
			{TagName: "v2.5.3"},
			{TagName: "v2.5.8"},
			{TagName: "v2.7.8"},
			{TagName: "v3.0.0"},
			{TagName: "v3.0.5"},
		}
		options := upgrade.NewReleaseOptions("", "v2.5", false)

		// Act
		release, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, &upgrade.Release{TagName: "v2.5.8"}, release)
	})

	t.Run("success_minor_option_newer", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{
			{TagName: "v1.7.8"},
			{TagName: "v2.3.8"},
			{TagName: "v2.5.3"},
			{TagName: "v2.5.8"},
			{TagName: "v2.7.8"},
			{TagName: "v3.0.0"},
			{TagName: "v3.0.5"},
		}
		options := upgrade.NewReleaseOptions("", "v2.5", false)

		// Act
		release, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, &upgrade.Release{TagName: "v2.5.8"}, release)
	})

	t.Run("success_prerelease_option", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{
			{TagName: "v1.6.7"},
			{TagName: "v2.3.8"},
			{TagName: "v3.0.5"},
			{TagName: "v4.5.7-beta.1"},
			{TagName: "v4.5.7"},
			{TagName: "v4.5.8-beta.2"},
		}
		options := upgrade.NewReleaseOptions("", "", true)

		// Act
		release, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, &upgrade.Release{TagName: "v4.5.8-beta.2"}, release)
	})

	t.Run("success_prerelease_option_but_latest", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{
			{TagName: "v1.6.7"},
			{TagName: "v2.3.8"},
			{TagName: "v3.0.5"},
			{TagName: "v4.5.7-beta.1"},
			{TagName: "v4.5.7"},
		}
		options := upgrade.NewReleaseOptions("", "", true)

		// Act
		release, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, &upgrade.Release{TagName: "v4.5.7"}, release)
	})

	t.Run("success_latest_same", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{
			{TagName: "v1.6.7"},
			{TagName: "v2.3.8"},
			{TagName: "v3.0.5"},
			{TagName: "v4.5.7"},
			{TagName: "v4.5.8-beta.1"},
		}
		options := upgrade.NewReleaseOptions("", "", false)

		// Act
		release, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, &upgrade.Release{TagName: "v4.5.7"}, release)
	})

	t.Run("success_latest_newer", func(t *testing.T) {
		// Arrange
		releases := []upgrade.Release{ // unordered slice of releases
			{TagName: "v4.7.3"},
			{TagName: "v3.0.5"},
			{TagName: "v1.6.7"},
			{TagName: "v4.5.8-beta.1"},
			{TagName: "v4.5.7"},
			{TagName: "v2.3.8"},
		}
		options := upgrade.NewReleaseOptions("", "", false)

		// Act
		release, ok := upgrade.FindRelease(releases, options)

		// Assert
		assert.True(t, ok)
		assert.Equal(t, &upgrade.Release{TagName: "v4.7.3"}, release)
	})
}

func TestGetDownloadURL(t *testing.T) {
	t.Run("error_no_valid_asset", func(t *testing.T) {
		// Arrange
		release := &upgrade.Release{}

		// Act
		_, err := upgrade.GetDownloadURL(release, "")

		// Assert
		assert.ErrorContains(t, err, "no valid release asset found")
	})

	t.Run("success_without_checksum", func(t *testing.T) {
		// Arrange
		release := &upgrade.Release{
			Assets: []upgrade.Asset{
				{DownloadURL: "zip URL", Name: fmt.Sprintf("%s_%s.zip", runtime.GOOS, runtime.GOARCH)},
				{DownloadURL: "tar.gz URL", Name: fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)},
				{DownloadURL: "deb URL", Name: fmt.Sprintf("%s_%s.deb", runtime.GOOS, runtime.GOARCH)},
				{DownloadURL: "apk URL", Name: fmt.Sprintf("%s_%s.apk", runtime.GOOS, runtime.GOARCH)},
			},
		}

		// Act
		url, err := upgrade.GetDownloadURL(release, fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH))

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "tar.gz URL", url)
	})

	t.Run("success_with_checksum", func(t *testing.T) {
		// Arrange
		release := &upgrade.Release{
			Assets: []upgrade.Asset{
				{DownloadURL: "apk URL", Name: fmt.Sprintf("%s_%s.apk", runtime.GOOS, runtime.GOARCH)},
				{DownloadURL: "checksum URL", Name: "checksums.txt"},
				{DownloadURL: "deb URL", Name: fmt.Sprintf("%s_%s.deb", runtime.GOOS, runtime.GOARCH)},
				{DownloadURL: "tar.gz URL", Name: fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)},
				{DownloadURL: "zip URL", Name: fmt.Sprintf("%s_%s.zip", runtime.GOOS, runtime.GOARCH)},
			},
		}

		// Act
		url, err := upgrade.GetDownloadURL(release, fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH))

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "tar.gz URL?checksum=file:checksum URL", url)
	})
}

func TestBinaryName(t *testing.T) {
	t.Run("success_major", func(t *testing.T) {
		// Arrange
		options := upgrade.NewReleaseOptions("v1", "", false)

		// Act
		name := upgrade.BinaryName("repo", "", "v1.5.6", options)

		// Assert
		assert.Equal(t, "repo-v1", name)
	})

	t.Run("success_major_prerelease", func(t *testing.T) {
		// Arrange
		options := upgrade.NewReleaseOptions("v1", "", true)

		// Act
		name := upgrade.BinaryName("repo", "", "v1.5.6-beta.1", options)

		// Assert
		assert.Equal(t, "repo-v1-pre", name)
	})

	t.Run("success_minor", func(t *testing.T) {
		// Arrange
		options := upgrade.NewReleaseOptions("", "v1.5", false)

		// Act
		name := upgrade.BinaryName("repo", "", "v1.5.6", options)

		// Assert
		assert.Equal(t, "repo-v1.5", name)
	})

	t.Run("success_minor_prerelease", func(t *testing.T) {
		// Arrange
		options := upgrade.NewReleaseOptions("", "v1.5", true)

		// Act
		name := upgrade.BinaryName("repo", "", "v1.5.6-beta.1", options)

		// Assert
		assert.Equal(t, "repo-v1.5-pre", name)
	})

	t.Run("success_prerelease", func(t *testing.T) {
		// Arrange
		options := upgrade.NewReleaseOptions("", "", true)

		// Act
		name := upgrade.BinaryName("repo", "", "v1.5.6-beta.1", options)

		// Assert
		assert.Equal(t, "repo-pre", name)
	})

	t.Run("success_latest_prerelease", func(t *testing.T) {
		// Arrange
		options := upgrade.NewReleaseOptions("", "", true)

		// Act
		name := upgrade.BinaryName("repo", ".sh", "v1.5.6", options)

		// Assert
		assert.Equal(t, "repo.sh", name)
	})

	t.Run("success_latest", func(t *testing.T) {
		// Arrange
		options := upgrade.NewReleaseOptions("", "", false)

		// Act
		name := upgrade.BinaryName("repo", ".exe", "v1.5.6", options)

		// Assert
		assert.Equal(t, "repo.exe", name)
	})
}
