package cli

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func versionString() string {
	return "yx version " + Version + " (" + Commit + ", " + Date + ")"
}
