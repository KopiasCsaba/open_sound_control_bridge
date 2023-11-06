package filetools

import (
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func MkdirIfNotExist(name string, perm os.FileMode) error {
	err := os.Mkdir(name, perm)
	if err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("unable to create directory: '%s': %w", name, err)
		}
	}
	return nil
}

func MkdirAllIfNotExist(name string, perm os.FileMode) error {
	err := os.MkdirAll(name, perm)
	if err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("unable to create directory: '%s': %w", name, err)
		}
	}
	return nil
}

func CopyFile(sourcePath string, destinationPath string) error {
	fInput, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open %s to copy to %s: %w", sourcePath, destinationPath, err)
	}
	defer func(fInput *os.File) {
		_ = fInput.Close()
	}(fInput)

	fOutput, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create %s to copy from %s: %w", destinationPath, sourcePath, err)
	}

	defer func(fOutput *os.File) {
		_ = fOutput.Close()
	}(fOutput)

	_, err = io.Copy(fOutput, fInput)
	if err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", sourcePath, destinationPath, err)
	}
	return nil
}

func FileNameWithoutExtension(fn string) string {
	return strings.TrimSuffix(fn, filepath.Ext(fn))
}

func FileBaseNameWithoutExtension(fn string) string {
	return strings.TrimSuffix(path.Base(fn), filepath.Ext(fn))
}

// CalculateHashOfFile sha256.New()
func CalculateHashOfFile(file string, hashAlgorithm hash.Hash) (string, error) {
	fileHandle, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", file, err)
	}
	defer fileHandle.Close()

	if _, err := io.Copy(hashAlgorithm, fileHandle); err != nil {
		return "", fmt.Errorf("failed to ophash file %s: %w", file, err)
	}
	sum := hashAlgorithm.Sum(nil)

	return fmt.Sprintf("%x", sum), nil
}
