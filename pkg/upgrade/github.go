package upgrade

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v63/github"
)

// GithubReleases returns a function listing all releases from a specific owner/repo in github.
//
// It's generic to help the reusability of this function when used as an SDK.
func GithubReleases(owner, repo string) func(ctx context.Context, httpClient *http.Client) ([]Release, error) {
	toReleases := func(releases []*github.RepositoryRelease) []Release {
		result := make([]Release, 0, len(releases))
		for _, release := range releases {
			if release == nil || release.TagName == nil {
				continue
			}
			r := Release{
				Assets:  make([]Asset, 0, len(release.Assets)),
				TagName: *release.TagName,
			}

			for _, asset := range release.Assets {
				if asset == nil || asset.Name == nil || asset.BrowserDownloadURL == nil {
					continue
				}
				r.Assets = append(r.Assets, Asset{DownloadURL: *asset.BrowserDownloadURL, Name: *asset.Name})
			}

			result = append(result, r)
		}
		return result
	}

	return func(ctx context.Context, httpClient *http.Client) ([]Release, error) {
		gCtx := context.WithValue(ctx, github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true) // handle github rate limiter

		client := github.NewClient(httpClient)
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			client = client.WithAuthToken(token)
		}

		var all []Release
		opts := github.ListOptions{PerPage: 100, Page: 1}
		for {
			releases, response, err := client.Repositories.ListReleases(gCtx, owner, repo, &opts)
			if err != nil {
				return nil, fmt.Errorf("list releases: %w", err)
			}
			all = append(all, toReleases(releases)...)

			if response.NextPage == 0 {
				break
			}
			opts.Page = response.NextPage
		}

		return all, nil
	}
}

var _ GetReleases = GithubReleases("owner", "repo") // ensure interface is implemented
