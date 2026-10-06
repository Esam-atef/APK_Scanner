package extractor

import (
	"archive/zip"
	"fmt"
	"io"
	"regexp"
)

const maxDexSize = 256 << 20

var dexNameRe = regexp.MustCompile(`^classes[0-9]*\.dex$`)

type APKInfo struct {
	Path  string
	Files []FileEntry
}

type FileEntry struct {
	Name             string 
	UncompressedSize uint64
}

func (a *APKInfo) Names() []string {
	names := make([]string, len(a.Files))
	for i, f := range a.Files {
		names[i] = f.Name
	}
	return names
}


func Open(path string) (*APKInfo, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file as a valid APK archive: %w", err)
	}
	defer r.Close()

	info := &APKInfo{Path: path}
	hasManifest := false

	for _, f := range r.File {
		info.Files = append(info.Files, FileEntry{
			Name:             f.Name,
			UncompressedSize: f.UncompressedSize64,
		})
		if f.Name == "AndroidManifest.xml" {
			hasManifest = true
		}
	}

	if !hasManifest {
		return nil, fmt.Errorf("file was opened as a ZIP but does not look like a valid APK (AndroidManifest.xml not found)")
	}

	return info, nil
}


func (a *APKInfo) ReadDexFiles(fn func(name string, data []byte, err error)) error {
	r, err := zip.OpenReader(a.Path)
	if err != nil {
		return fmt.Errorf("failed to re-open APK for DEX scanning: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if !dexNameRe.MatchString(f.Name) {
			continue
		}
		data, err := readEntry(f)
		fn(f.Name, data, err)
	}
	return nil
}

func readEntry(f *zip.File) ([]byte, error) {
	if f.UncompressedSize64 > maxDexSize {
		return nil, fmt.Errorf("file is larger than the %d MiB scan limit", maxDexSize>>20)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, maxDexSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxDexSize {
		return nil, fmt.Errorf("file is larger than the %d MiB scan limit", maxDexSize>>20)
	}
	return data, nil
}
