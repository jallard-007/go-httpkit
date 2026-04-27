package precompressed

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/jallard-007/go-httpkit/negotiate"
)

type FsInfo struct {
	EncodingToExt map[string]string
	Paths         map[string]FileInfo
}

type FileInfo struct {
	ContentType        string
	AvailableEncodings []string // ordered by server preference
}

func NewFsInfo(pathVariants map[string]Variants, encodingPref []string, extToEncoding map[string]string) FsInfo {
	extToEncoding[""] = "identity"
	encodingToExt := make(map[string]string, len(extToEncoding))
	for ex, en := range extToEncoding {
		encodingToExt[en] = ex
	}

	fsInfo := FsInfo{
		EncodingToExt: encodingToExt,
		Paths:         make(map[string]FileInfo, len(pathVariants)),
	}

	for p, vs := range pathVariants {
		var ae []string
		for _, e := range encodingPref {
			_, ok := vs[e]
			if ok {
				ae = append(ae, e)
			}
		}

		ext := strings.ToLower(filepath.Ext(p))
		mimeType := mime.TypeByExtension(ext)
		fsInfo.Paths[p] = FileInfo{
			ContentType:        mimeType,
			AvailableEncodings: ae,
		}
	}

	return fsInfo
}

func (fsys *FsInfo) Validate() error {
	for f, i := range fsys.Paths {
		if len(i.AvailableEncodings) == 0 {
			return fmt.Errorf("file %q has no available content encodings. at least one should be available, typically \"identity\"", f)
		}
		for _, e := range i.AvailableEncodings {
			if _, ok := fsys.EncodingToExt[e]; !ok {
				return fmt.Errorf("file %q's encoding of %q has no extension mapping", f, e)
			}
		}
	}
	return nil
}

func Handler(fsys fs.FS, fsysInfo *FsInfo) http.Handler {
	if err := fsysInfo.Validate(); err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := strings.TrimLeft(r.URL.Path, "/")

		i, ok := fsysInfo.Paths[urlPath]
		if !ok {
			encoding := negotiate.EncodingH(r.Header, nil)
			if encoding == "" {
				http.Error(w, "only \"identity\" content encoding is supported", http.StatusNotAcceptable)
			} else {
				http.ServeFileFS(w, r, fsys, urlPath)
			}
			return
		}

		encoding := negotiate.EncodingH(r.Header, i.AvailableEncodings)
		if encoding == "" {
			http.Error(w, fmt.Sprintf("only %v content encodings are supported", i.AvailableEncodings), http.StatusNotAcceptable)
			return
		}

		wh := w.Header()
		if i.ContentType != "" {
			wh["Content-Type"] = []string{i.ContentType}
		}

		ext := fsysInfo.EncodingToExt[encoding]
		p := urlPath + ext
		wh["Content-Encoding"] = []string{encoding}
		if len(i.AvailableEncodings) > 1 {
			wh["Vary"] = append(wh["Vary"], "Accept-Encoding")
		}

		http.ServeFileFS(w, r, fsys, p)
	})
}
