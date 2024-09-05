package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Eyepan/yap/src/types"
)

func ReadLock() (*types.Lockfile, error) {
	lockFilePath := filepath.Join(".", "yap.lockb")

	if _, err := os.Stat(lockFilePath); err == nil {
		return nil, fmt.Errorf("something went wrong while reading the lockfile: %w", err)
	}

	data, err := os.ReadFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while reading the lockfile: %w", err)

	}
	buf := bytes.NewReader(data)
	lockBin, err := ReadLockfile(buf)
	if err != nil {
		return nil, fmt.Errorf("something went wrong while reading the lockfile: %w", err)
	}
	return lockBin, nil
}

func DoesLockfileExist() (bool, error) {
	lockFile := filepath.Join(".", "yap.lockb")

	if _, err := os.Stat(lockFile); err != nil {
		return false, fmt.Errorf("something went wrong while reading the lockfile: %w", err)
	}
	return true, nil
}

func WriteLock(lockBin types.Lockfile) error {
	lockFilePath := filepath.Join(".", "yap.lockb")

	var buf bytes.Buffer
	if err := WriteLockfile(&buf, lockBin); err != nil {
		return fmt.Errorf("failed to write lockfile: %w", err)
	}
	file, err := os.Create(lockFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(buf.Bytes()); err != nil {
		return err
	}
	return nil
}
