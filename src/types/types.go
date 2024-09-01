package types

// Dependency represents a package dependency.
type Dependency struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MetadataDist contains distribution information for a package.
type MetadataDist struct {
	Shasum    string `json:"shasum"`
	Tarball   string `json:"tarball"`
	FileCount int64  `json:"fileCount"`
}

// VersionMetadata holds version-specific metadata for a package.
type VersionMetadata struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	ID           string            `json:"id"`
	Dist         MetadataDist      `json:"dist"`
	Dependencies map[string]string `json:"dependencies"`
}

// Tags represents the different tags associated with a package version.
type Tags struct {
	Latest string `json:"latest"`
	Next   string `json:"next"`
}

// Metadata represents the metadata for a package including its versions.
type Metadata struct {
	Name     string                     `json:"name"`
	ID       string                     `json:"_id"`
	DistTags Tags                       `json:"dist-tags"`
	Versions map[string]VersionMetadata `json:"versions"`
}

// Config is a map of configuration key-value pairs.
type Config map[string]string

// PackageJSON represents the structure of a package.json file.
type PackageJSON struct {
	Name             string            `json:"name"`
	Module           string            `json:"module"`
	Type             string            `json:"type"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Dependencies     map[string]string `json:"dependencies"`
}
