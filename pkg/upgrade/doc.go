/*
The upgrade package provides the possibility to upgrade / install any package with various tunings:
- A specific major version
- A specific minor version
- Include prereleases
- Installation destination
- Uninstall older versions

Note than when using major or minor options, the current version will not be used.
Why ? Because one could want to install an older version in case a breaking change was made by error
or for any other reason.

Example:

	func main() {
		ctx := context.Background()
		currentVersion := "v1.0.0"

		err := upgrade.Run(ctx, currentVersion, upgrade.GithubReleases("owner", "repo"), // where to retrieve releases
			upgrade.WithDestination("/tmp"), // installation destination
			upgrade.WithKeepVersions(true), // whether to keep older versions or not
			upgrade.WithMajor(""), // whether to install a specific major version or not
			upgrade.WithMinor(""), // whether to install a specific minor version or not
			upgrade.WithPrerelease(false), // whether to include prereleases in filtering
		)
	}
*/
package upgrade
