package upgrade_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v63/github"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"

	"github.com/kilianpaquier/cli-sdk/pkg/upgrade"
)

func TestGithubReleases(t *testing.T) {
	ctx := context.Background()

	// setup github / go-getter mocking
	httpClient := cleanhttp.DefaultClient()
	httpmock.ActivateNonDefault(httpClient)
	t.Cleanup(httpmock.DeactivateAndReset)

	getReleases := upgrade.GithubReleases("owner", "repo")

	t.Run("error_pagination", func(t *testing.T) {
		// Arrange
		releasesURL := "https://api.github.com/repos/owner/repo/releases"
		httpmock.RegisterResponder(http.MethodGet, releasesURL,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{{TagName: toPtr("v1.0.0")}}).
				HeaderAdd(map[string][]string{
					"Link": {fmt.Sprintf(`<%s?page=2&per_page=100>; rel="next"`, releasesURL)},
				}).
				Then(httpmock.NewStringResponder(http.StatusInternalServerError, "error message")))

		// Act
		_, err := getReleases(ctx, httpClient)

		// Assert
		assert.ErrorContains(t, err, "page=2&per_page=100") // error happened on page 2
	})

	t.Run("success_multiple_pages", func(t *testing.T) {
		// Arrange
		releasesURL := "https://api.github.com/repos/owner/repo/releases"
		httpmock.RegisterResponder(http.MethodGet, releasesURL,
			httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{{TagName: toPtr("v1.0.0")}}).
				HeaderAdd(map[string][]string{
					"Link": {fmt.Sprintf(`<%s?page=2&per_page=100>; rel="next"`, releasesURL)},
				}).
				Then(httpmock.NewJsonResponderOrPanic(http.StatusOK, []*github.RepositoryRelease{{TagName: toPtr("v1.0.1")}})))

		// Act
		releases, err := getReleases(ctx, httpClient)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, []upgrade.Release{
			{TagName: "v1.0.0", Assets: []upgrade.Asset{}},
			{TagName: "v1.0.1", Assets: []upgrade.Asset{}},
		}, releases)
	})
}
