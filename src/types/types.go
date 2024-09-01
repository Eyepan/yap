package types

type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type MetadataDist struct {
	Shasum    string `json:"shasum"`
	Tarball   string `json:"tarball"`
	FileCount int64  `json:"fileCount"`
}

type VersionMetadata struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	ID           string            `json:"id"`
	Dist         MetadataDist      `json:"dist"`
	Dependencies map[string]string `json:"dependencies"`
}

type Tags struct {
	Latest string `json:"latest"`
	Next   string `json:"next"`
}

type Metadata struct {
	Name     string                     `json:"name"`
	ID       string                     `json:"_id"`
	DistTags Tags                       `json:"dist-tags"`
	Versions map[string]VersionMetadata `json:"versions"`
}

type Config map[string]string

type PackageJSON struct {
	Name             string            `json:"name"`
	Module           string            `json:"module"`
	Type             string            `json:"type"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Dependencies     map[string]string `json:"dependencies"`
}
