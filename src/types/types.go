package types

type Config map[string]string

type Dependencies map[string]string

type PackageJSON struct {
	Name             string       `json:"name"`
	Module           string       `json:"module"`
	Type             string       `json:"type"`
	DevDependencies  Dependencies `json:"devDependencies"`
	PeerDependencies Dependencies `json:"peerDependencies"`
	Dependencies     Dependencies `json:"dependencies"`
}

type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type MPackage struct {
	Name         string
	Version      string
	Dist         Dist
	Dependencies []*MPackage
}

type Dist struct {
	Shasum    string `json:"shasum"`
	Tarball   string `json:"tarball"`
	FileCount int64  `json:"fileCount"`
}

type Lockfile struct {
	CoreDependencies []Package
	Resolutions      []MPackage
}

type Metadata struct {
	Name     string `json:"name"`
	DistTags struct {
		Latest string `json:"latest"`
		Next   string `json:"next"`
	} `json:"dist-tags"`
	Versions map[string]VersionMetadata `json:"versions"`
}

type VersionMetadata struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	Dist         Dist         `json:"dist"`
	Dependencies Dependencies `json:"dependencies"`
}
