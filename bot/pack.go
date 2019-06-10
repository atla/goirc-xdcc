package bot

// Package information to trigger dcc send
type Package struct {
	Host             string
	Network          string
	Channel          string
	CompanionChannel string
	PackageID        int
}
