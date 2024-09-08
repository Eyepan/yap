package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/Eyepan/yap/src/types"
)

func writeString(buf *bytes.Buffer, str string) error {
	if err := binary.Write(buf, binary.LittleEndian, int32(len(str))); err != nil {
		return fmt.Errorf("failed to write string length to buffer: %w", err)
	}
	if _, err := buf.WriteString(str); err != nil {
		return fmt.Errorf("failed to write string content to buffer: %w", err)
	}
	return nil
}

func writeVersionMetadata(buf *bytes.Buffer, vm types.VersionMetadata) error {
	if err := writeString(buf, vm.Name); err != nil {
		return fmt.Errorf("failed to write version metadata name: %w", err)
	}
	if err := writeString(buf, vm.Version); err != nil {
		return fmt.Errorf("failed to write version metadata version: %w", err)
	}
	if err := writeString(buf, vm.Dist.Shasum); err != nil {
		return fmt.Errorf("failed to write version metadata shasum: %w", err)
	}
	if err := writeString(buf, vm.Dist.Tarball); err != nil {
		return fmt.Errorf("failed to write version metadata tarball: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, vm.Dist.FileCount); err != nil {
		return fmt.Errorf("failed to write version metadata file count: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, int32(len(vm.Dependencies))); err != nil {
		return fmt.Errorf("failed to write dependencies count: %w", err)
	}
	for k, v := range vm.Dependencies {
		if err := writeString(buf, k); err != nil {
			return fmt.Errorf("failed to write dependency key: %w", err)
		}
		if err := writeString(buf, v); err != nil {
			return fmt.Errorf("failed to write dependency value: %w", err)
		}
	}

	return nil
}

func WriteMetadata(buf *bytes.Buffer, metadata types.Metadata) error {
	if err := writeString(buf, metadata.Name); err != nil {
		return fmt.Errorf("failed to write metadata name: %w", err)
	}
	if err := writeString(buf, metadata.DistTags.Latest); err != nil {
		return fmt.Errorf("failed to write metadata dist tag latest: %w", err)
	}
	if err := writeString(buf, metadata.DistTags.Next); err != nil {
		return fmt.Errorf("failed to write metadata dist tag next: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, int32(len(metadata.Versions))); err != nil {
		return fmt.Errorf("failed to write versions count: %w", err)
	}
	for k, v := range metadata.Versions {
		if err := writeString(buf, k); err != nil {
			return fmt.Errorf("failed to write version key: %w", err)
		}
		if err := writeVersionMetadata(buf, v); err != nil {
			return fmt.Errorf("failed to write version metadata: %w", err)
		}
	}

	return nil
}

func readString(buf *bytes.Reader) (string, error) {
	var length int32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}

	strBytes := make([]byte, length)
	if _, err := buf.Read(strBytes); err != nil {
		return "", fmt.Errorf("failed to read string content: %w", err)
	}

	return string(strBytes), nil
}

func readVersionMetadata(buf *bytes.Reader) (types.VersionMetadata, error) {
	var vm types.VersionMetadata
	var err error
	if vm.Name, err = readString(buf); err != nil {
		return vm, fmt.Errorf("failed to read version metadata name: %w", err)
	}
	if vm.Version, err = readString(buf); err != nil {
		return vm, fmt.Errorf("failed to read version metadata version: %w", err)
	}
	if vm.Dist.Shasum, err = readString(buf); err != nil {
		return vm, fmt.Errorf("failed to read version metadata shasum: %w", err)
	}
	if vm.Dist.Tarball, err = readString(buf); err != nil {
		return vm, fmt.Errorf("failed to read version metadata tarball: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &vm.Dist.FileCount); err != nil {
		return vm, fmt.Errorf("failed to read version metadata file count: %w", err)
	}

	var depCount int32
	if err := binary.Read(buf, binary.LittleEndian, &depCount); err != nil {
		return vm, fmt.Errorf("failed to read dependencies count: %w", err)
	}
	vm.Dependencies = make(map[string]string, depCount)
	for i := 0; i < int(depCount); i++ {
		key, err := readString(buf)
		if err != nil {
			return vm, fmt.Errorf("failed to read dependency key: %w", err)
		}
		value, err := readString(buf)
		if err != nil {
			return vm, fmt.Errorf("failed to read dependency value: %w", err)
		}
		vm.Dependencies[key] = value
	}

	return vm, nil
}

func ReadMetadata(buf *bytes.Reader) (*types.Metadata, error) {
	var metadata types.Metadata
	var err error
	if metadata.Name, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read metadata name: %w", err)
	}
	if metadata.DistTags.Latest, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read metadata dist tag latest: %w", err)
	}
	if metadata.DistTags.Next, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read metadata dist tag next: %w", err)
	}

	var versionCount int32
	if err := binary.Read(buf, binary.LittleEndian, &versionCount); err != nil {
		return nil, fmt.Errorf("failed to read versions count: %w", err)
	}
	metadata.Versions = make(map[string]types.VersionMetadata, versionCount)
	for i := 0; i < int(versionCount); i++ {
		key, err := readString(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read version key: %w", err)
		}
		vm, err := readVersionMetadata(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read version metadata: %w", err)
		}
		metadata.Versions[key] = vm
	}

	return &metadata, nil
}

func writePackage(buf *bytes.Buffer, pkg types.Package) error {
	if err := writeString(buf, pkg.Name); err != nil {
		return fmt.Errorf("failed to write package name: %w", err)
	}
	if err := writeString(buf, pkg.Version); err != nil {
		return fmt.Errorf("failed to write package version: %w", err)
	}
	return nil
}

func writeMPackage(buf *bytes.Buffer, mPackage *types.MPackage) error {
	if err := writeString(buf, mPackage.Name); err != nil {
		return fmt.Errorf("failed to write mPackage name: %w", err)
	}
	if err := writeString(buf, mPackage.Version); err != nil {
		return fmt.Errorf("failed to write mPackage version: %w", err)
	}
	if err := writeString(buf, mPackage.Dist.Shasum); err != nil {
		return fmt.Errorf("failed to write mPackage shasum: %w", err)
	}
	if err := writeString(buf, mPackage.Dist.Tarball); err != nil {
		return fmt.Errorf("failed to write mPackage tarball: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, mPackage.Dist.FileCount); err != nil {
		return fmt.Errorf("failed to write mPackage file count: %w", err)
	}

	if err := binary.Write(buf, binary.LittleEndian, int32(len(mPackage.Dependencies))); err != nil {
		return fmt.Errorf("failed to write mPackage dependencies count: %w", err)
	}
	for _, dep := range mPackage.Dependencies {
		if err := writeMPackage(buf, dep); err != nil {
			return fmt.Errorf("failed to write mPackage dependency: %w", err)
		}
	}

	return nil
}

func WriteLockfile(buf *bytes.Buffer, lockfile types.Lockfile) error {
	if err := binary.Write(buf, binary.LittleEndian, int32(len(lockfile.CoreDependencies))); err != nil {
		return fmt.Errorf("failed to write core dependencies count: %w", err)
	}
	for _, pkg := range lockfile.CoreDependencies {
		if err := writePackage(buf, pkg); err != nil {
			return fmt.Errorf("failed to write core dependency package: %w", err)
		}
	}

	if err := binary.Write(buf, binary.LittleEndian, int32(len(lockfile.Resolutions))); err != nil {
		return fmt.Errorf("failed to write resolutions count: %w", err)
	}
	for _, mPackage := range lockfile.Resolutions {
		if err := writeMPackage(buf, &mPackage); err != nil {
			return fmt.Errorf("failed to write resolution mPackage: %w", err)
		}
	}

	return nil
}

func readPackage(buf *bytes.Reader) (types.Package, error) {
	var pkg types.Package

	var err error
	if pkg.Name, err = readString(buf); err != nil {
		return pkg, fmt.Errorf("failed to read package name: %w", err)
	}
	if pkg.Version, err = readString(buf); err != nil {
		return pkg, fmt.Errorf("failed to read package version: %w", err)
	}

	return pkg, nil
}

func readMPackage(buf *bytes.Reader) (*types.MPackage, error) {
	var mPackage types.MPackage

	var err error
	if mPackage.Name, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read mPackage name: %w", err)
	}
	if mPackage.Version, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read mPackage version: %w", err)
	}
	if mPackage.Dist.Shasum, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read mPackage shasum: %w", err)
	}
	if mPackage.Dist.Tarball, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read mPackage tarball: %w", err)
	}
	if err := binary.Read(buf, binary.LittleEndian, &mPackage.Dist.FileCount); err != nil {
		return nil, fmt.Errorf("failed to read mPackage file count: %w", err)
	}

	var depCount int32
	if err := binary.Read(buf, binary.LittleEndian, &depCount); err != nil {
		return nil, fmt.Errorf("failed to read mPackage dependencies count: %w", err)
	}
	mPackage.Dependencies = make([]*types.MPackage, depCount)
	for i := 0; i < int(depCount); i++ {
		dep, err := readMPackage(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read mPackage dependency: %w", err)
		}
		mPackage.Dependencies[i] = dep
	}

	return &mPackage, nil
}

func ReadLockfile(buf *bytes.Reader) (*types.Lockfile, error) {
	var lockfile types.Lockfile

	var err error
	var coreDepCount int32
	if err = binary.Read(buf, binary.LittleEndian, &coreDepCount); err != nil {
		return nil, fmt.Errorf("failed to read core dependencies count: %w", err)
	}
	lockfile.CoreDependencies = make([]types.Package, coreDepCount)
	for i := 0; i < int(coreDepCount); i++ {
		if lockfile.CoreDependencies[i], err = readPackage(buf); err != nil {
			return nil, fmt.Errorf("failed to read core dependency package: %w", err)
		}
	}

	var resCount int32
	if err = binary.Read(buf, binary.LittleEndian, &resCount); err != nil {
		return nil, fmt.Errorf("failed to read resolutions count: %w", err)
	}
	lockfile.Resolutions = make([]types.MPackage, resCount)
	for i := 0; i < int(resCount); i++ {
		mPkg, err := readMPackage(buf)
		if err != nil {
			return nil, fmt.Errorf("failed to read resolution mPackage: %w", err)
		}
		lockfile.Resolutions[i] = *mPkg
	}

	return &lockfile, nil
}

func WriteConfig(buf *bytes.Buffer, conf *types.YapConfig) error {
	if err := writeString(buf, conf.Registry); err != nil {
		return fmt.Errorf("failed to write config registry: %w", err)
	}
	if err := writeString(buf, conf.AuthToken); err != nil {
		return fmt.Errorf("failed to write config auth token: %w", err)
	}
	if err := writeString(buf, string(conf.LogLevel)); err != nil {
		return fmt.Errorf("failed to write config log level: %w", err)
	}
	return nil
}

func ReadConfig(buf *bytes.Reader) (*types.YapConfig, error) {
	var conf types.YapConfig

	var err error
	if conf.Registry, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read config registry: %w", err)
	}
	if conf.AuthToken, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read config auth token: %w", err)
	}
	if conf.LogLevel, err = readString(buf); err != nil {
		return nil, fmt.Errorf("failed to read config log level: %w", err)
	}

	return &conf, nil
}
