package tools

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

func copyUploads(source, destination string) ([]UploadChecksum, error) {
	info, err := os.Lstat(source)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(destination, 0o700); err != nil {
			return nil, err
		}
		return []UploadChecksum{}, nil
	}
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("upload root is a symlink")
	}
	if !info.IsDir() {
		return nil, errors.New("upload root is not a directory")
	}
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return nil, err
	}
	checksums := make([]UploadChecksum, 0)
	err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		checksum, err := copyUploadFile(path, target, filepath.ToSlash(relative))
		if err != nil {
			return err
		}
		checksums = append(checksums, checksum)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(checksums, func(i, j int) bool { return checksums[i].Path < checksums[j].Path })
	return checksums, nil
}

func copyUploadFile(source, destination, relative string) (checksum UploadChecksum, err error) {
	info, err := os.Lstat(source)
	if err != nil {
		return checksum, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return checksum, errors.New("upload changed to a symlink")
	}
	input, err := os.Open(source)
	if err != nil {
		return checksum, err
	}
	defer func() { err = errors.Join(err, input.Close()) }()
	current, err := os.Lstat(source)
	if err != nil {
		return checksum, err
	}
	if current.Mode()&os.ModeSymlink != 0 || !os.SameFile(info, current) {
		return checksum, errors.New("upload changed while it was being copied")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return checksum, err
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm()&0o666)
	if err != nil {
		return checksum, err
	}
	defer func() { err = errors.Join(err, output.Close()) }()
	hasher := sha256.New()
	bytesCopied, err := io.Copy(io.MultiWriter(output, hasher), input)
	if err != nil {
		return checksum, err
	}
	if err := output.Sync(); err != nil {
		return checksum, err
	}
	checksum = UploadChecksum{Path: relative, SHA256: hex.EncodeToString(hasher.Sum(nil)), Bytes: bytesCopied}
	return checksum, nil
}

func hashFile(path string) (hash string, err error) {
	input, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, input.Close()) }()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, input); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
