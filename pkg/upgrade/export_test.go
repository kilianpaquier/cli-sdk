package upgrade

var (
	BinaryName     = binaryName
	FindRelease    = findRelease
	GetDownloadURL = getDownloadURL
)

func NewReleaseOptions(major, minor string, prerelease bool) releaseOptions { // nolint:revive
	return releaseOptions{
		major:      major,
		minor:      minor,
		prerelease: prerelease,
	}
}
