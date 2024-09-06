package utils

import (
	"bytes"
	"encoding/binary"

	"github.com/Eyepan/yap/src/types"
)

func writeString(buf *bytes.Buffer, str string) error {
	if err := binary.Write(buf, binary.LittleEndian, int32(len(str))); err != nil {
		return err
	}
	if _, err := buf.WriteString(str); err != nil {
		return err
	}
	return nil
}

func writeVersionMetadata(buf *bytes.Buffer, vm types.VersionMetadata) error {
	if err := writeString(buf, vm.Name); err != nil {
		return err
	}
	if err := writeString(buf, vm.Version); err != nil {
		return err
	}
	if err := writeString(buf, vm.Dist.Shasum); err != nil {
		return err
	}
	if err := writeString(buf, vm.Dist.Tarball); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, vm.Dist.FileCount); err != nil {
		return err
	}

	// Write dependencies map
	if err := binary.Write(buf, binary.LittleEndian, int32(len(vm.Dependencies))); err != nil {
		return err
	}
	for k, v := range vm.Dependencies {
		if err := writeString(buf, k); err != nil {
			return err
		}
		if err := writeString(buf, v); err != nil {
			return err
		}
	}

	return nil
}

func WriteMetadata(buf *bytes.Buffer, metadata types.Metadata) error {
	if err := writeString(buf, metadata.Name); err != nil {
		return err
	}
	if err := writeString(buf, metadata.DistTags.Latest); err != nil {
		return err
	}
	if err := writeString(buf, metadata.DistTags.Next); err != nil {
		return err
	}

	// Write versions map
	if err := binary.Write(buf, binary.LittleEndian, int32(len(metadata.Versions))); err != nil {
		return err
	}
	for k, v := range metadata.Versions {
		if err := writeString(buf, k); err != nil {
			return err
		}
		if err := writeVersionMetadata(buf, v); err != nil {
			return err
		}
	}

	return nil
}

func readString(buf *bytes.Reader) (string, error) {
	var length int32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", err
	}

	strBytes := make([]byte, length)
	if _, err := buf.Read(strBytes); err != nil {
		return "", err
	}

	return string(strBytes), nil
}

func readVersionMetadata(buf *bytes.Reader) (types.VersionMetadata, error) {
	var vm types.VersionMetadata
	var err error

	if vm.Name, err = readString(buf); err != nil {
		return vm, err
	}
	if vm.Version, err = readString(buf); err != nil {
		return vm, err
	}
	if vm.Dist.Shasum, err = readString(buf); err != nil {
		return vm, err
	}
	if vm.Dist.Tarball, err = readString(buf); err != nil {
		return vm, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &vm.Dist.FileCount); err != nil {
		return vm, err
	}

	var depCount int32
	if err := binary.Read(buf, binary.LittleEndian, &depCount); err != nil {
		return vm, err
	}
	vm.Dependencies = make(map[string]string, depCount)
	for i := 0; i < int(depCount); i++ {
		key, err := readString(buf)
		if err != nil {
			return vm, err
		}
		value, err := readString(buf)
		if err != nil {
			return vm, err
		}
		vm.Dependencies[key] = value
	}

	return vm, nil
}

func ReadMetadata(buf *bytes.Reader) (*types.Metadata, error) {
	var metadata types.Metadata
	var err error

	if metadata.Name, err = readString(buf); err != nil {
		return nil, err
	}
	if metadata.DistTags.Latest, err = readString(buf); err != nil {
		return nil, err
	}
	if metadata.DistTags.Next, err = readString(buf); err != nil {
		return nil, err
	}

	var versionCount int32
	if err := binary.Read(buf, binary.LittleEndian, &versionCount); err != nil {
		return nil, err
	}
	metadata.Versions = make(map[string]types.VersionMetadata, versionCount)
	for i := 0; i < int(versionCount); i++ {
		key, err := readString(buf)
		if err != nil {
			return nil, err
		}
		vm, err := readVersionMetadata(buf)
		if err != nil {
			return nil, err
		}
		metadata.Versions[key] = vm
	}

	return &metadata, nil
}

func writePackage(buf *bytes.Buffer, pkg types.Package) error {
	if err := writeString(buf, pkg.Name); err != nil {
		return err
	}
	if err := writeString(buf, pkg.Version); err != nil {
		return err
	}
	return nil
}

func writeMPackage(buf *bytes.Buffer, mPackage *types.MPackage) error {
	if err := writeString(buf, mPackage.Name); err != nil {
		return err
	}
	if err := writeString(buf, mPackage.Version); err != nil {
		return err
	}
	if err := writeString(buf, mPackage.Dist.Shasum); err != nil {
		return err
	}
	if err := writeString(buf, mPackage.Dist.Tarball); err != nil {
		return err
	}
	if err := binary.Write(buf, binary.LittleEndian, mPackage.Dist.FileCount); err != nil {
		return err
	}

	// Write Dependencies recursively
	if err := binary.Write(buf, binary.LittleEndian, int32(len(mPackage.Dependencies))); err != nil {
		return err
	}
	for _, dep := range mPackage.Dependencies {
		if err := writeMPackage(buf, dep); err != nil {
			return err
		}
	}

	return nil
}

func WriteLockfile(buf *bytes.Buffer, lockfile types.Lockfile) error {
	// Write CoreDependencies
	if err := binary.Write(buf, binary.LittleEndian, int32(len(lockfile.CoreDependencies))); err != nil {
		return err
	}
	for _, pkg := range lockfile.CoreDependencies {
		if err := writePackage(buf, pkg); err != nil {
			return err
		}
	}

	// Write Resolutions
	if err := binary.Write(buf, binary.LittleEndian, int32(len(lockfile.Resolutions))); err != nil {
		return err
	}
	for _, mPackage := range lockfile.Resolutions {
		if err := writeMPackage(buf, &mPackage); err != nil {
			return err
		}
	}

	return nil
}

func readPackage(buf *bytes.Reader) (types.Package, error) {
	var pkg types.Package
	var err error

	if pkg.Name, err = readString(buf); err != nil {
		return pkg, err
	}
	if pkg.Version, err = readString(buf); err != nil {
		return pkg, err
	}

	return pkg, nil
}

func readMPackage(buf *bytes.Reader) (*types.MPackage, error) {
	var mPackage types.MPackage
	var err error

	if mPackage.Name, err = readString(buf); err != nil {
		return nil, err
	}
	if mPackage.Version, err = readString(buf); err != nil {
		return nil, err
	}
	if mPackage.Dist.Shasum, err = readString(buf); err != nil {
		return nil, err
	}
	if mPackage.Dist.Tarball, err = readString(buf); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &mPackage.Dist.FileCount); err != nil {
		return nil, err
	}

	// Read Dependencies recursively
	var depCount int32
	if err := binary.Read(buf, binary.LittleEndian, &depCount); err != nil {
		return nil, err
	}
	mPackage.Dependencies = make([]*types.MPackage, depCount)
	for i := 0; i < int(depCount); i++ {
		dep, err := readMPackage(buf)
		if err != nil {
			return nil, err
		}
		mPackage.Dependencies[i] = dep
	}

	return &mPackage, nil
}

func ReadLockfile(buf *bytes.Reader) (*types.Lockfile, error) {
	var lockfile types.Lockfile

	// Read CoreDependencies
	var coreDepCount int32
	if err := binary.Read(buf, binary.LittleEndian, &coreDepCount); err != nil {
		return nil, err
	}
	lockfile.CoreDependencies = make([]types.Package, coreDepCount)
	for i := 0; i < int(coreDepCount); i++ {
		pkg, err := readPackage(buf)
		if err != nil {
			return nil, err
		}
		lockfile.CoreDependencies[i] = pkg
	}

	// Read Resolutions
	var resCount int32
	if err := binary.Read(buf, binary.LittleEndian, &resCount); err != nil {
		return nil, err
	}
	lockfile.Resolutions = make([]types.MPackage, resCount)
	for i := 0; i < int(resCount); i++ {
		mPackage, err := readMPackage(buf)
		if err != nil {
			return nil, err
		}
		lockfile.Resolutions[i] = *mPackage
	}

	return &lockfile, nil
}
