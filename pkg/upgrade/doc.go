/*
The upgrade package provides the possibility to upgrade / install any package with various tunings:

  - Specify the asset name to download (with templating)
  - Installation destination
  - Specify the target binary name (with templating)
  - A specific major version
  - A specific minor version
  - Include prereleases

Note than when using major or minor options, the current version will not be used.
Why ? Because one could want to install an older version in case a breaking change was made by error
or for any other reason.

Example:

	func main() {
		ctx := context.Background()

		// project name, used to define the final binary name
		// suffixed by '-' and the major version in case it's given with WithMajor
		// suffixed by '-' and the minor version in case it's given with WithMinor
		// suffixed by '-beta' or '-alpha' or '-pre', etc. in case prereleases option is given with WithPrereleases
		// and the installed release is in fact a prerelease
		//
		// still, can be tuned with upgrade.WithTargetTemplate function (see associated doc)
		repo := "repo"

		currentVersion := "v1.0.0" // currently installed version

		err := upgrade.Run(ctx, repo, currentVersion,
			upgrade.GithubReleases("owner", repo), // where to retrieve releases
			upgrade.WithAssetTemplate("{{ .Repo }}_{{ .GOOS }}_{{ .GOARCH }}.tar.gz"), // which asset should be downloaded (see associated doc)
			upgrade.WithDestination("/tmp"), // installation destination
			upgrade.WithKeepVersions(true), // whether to keep older versions or not
			upgrade.WithMajor(""), // whether to install a specific major version or not
			upgrade.WithMinor(""), // whether to install a specific minor version or not
			upgrade.WithPrerelease(false), // whether to include prereleases in filtering
		)
	}
*/
package upgrade
