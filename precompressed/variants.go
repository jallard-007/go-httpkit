package precompressed

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// set of encoding variants of a file
// e.g. [br, gzip]
type Variants map[string]struct{}

func GatherVariants(fsys fs.FS, extToEncoding map[string]string) map[string]Variants {
	paths := gatherPaths(fsys)
	files := make(map[string]Variants, len(paths))
	for path := range paths {
		ext := strings.ToLower(filepath.Ext(path))
		pWithoutExt := path[:len(path)-len(ext)]
		_, ok := paths[pWithoutExt]
		if !ok {
			files[path] = Variants{"identity": struct{}{}}
			continue
		}
	}

	for path := range paths {
		ext := strings.ToLower(filepath.Ext(path))
		pWithoutExt := path[:len(path)-len(ext)]
		_, ok := paths[pWithoutExt]
		if !ok {
			continue
		}

		n, ok := extToEncoding[ext]
		if !ok {
			continue
		}

		v, ok := files[pWithoutExt]
		if !ok {
			continue
		}
		if v == nil {
			v = make(map[string]struct{})
		}
		v[n] = struct{}{}
		files[pWithoutExt] = v
	}

	return files
}

func gatherPaths(fsys fs.FS) map[string]struct{} {
	paths := make(map[string]struct{})
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		paths[path] = struct{}{}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return paths
}
