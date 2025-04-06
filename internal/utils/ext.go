package utils

import (
	"path/filepath"
	"strings"
)

func GetFullFileExtension(path string) string {
	ext := filepath.Ext(path)
	name := strings.TrimSuffix(path, ext)

	prevExt := filepath.Ext(name)
	if prevExt != "" {
		return prevExt + ext
	}
	return ext
}

func RemoveFileExtension(path string) string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	secondExt := filepath.Ext(name)
	if secondExt != "" && ext == ".gz" {
		name = strings.TrimSuffix(name, secondExt)
	}

	return filepath.Join(dir, name)
}
