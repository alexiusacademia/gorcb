package version

// These variables are set at build time using -ldflags
// Example: go build -ldflags "-X github.com/alexiusacademia/gorcb/internal/version.Version=1.0.0"
var (
	// Version is the semantic version of the application
	Version = "1.0.1"

	// BuildTime is the time the binary was built (set via ldflags)
	BuildTime = "unknown"

	// GitCommit is the git commit hash (set via ldflags)
	GitCommit = "unknown"

	// Author of the application
	Author = "Alexius Academia"

	// Year of release
	Year = "2025"
)

