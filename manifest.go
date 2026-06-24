package vite

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
)

// manifestEntry is one record in Vite's manifest.json. Only the fields used for
// backend asset resolution are decoded.
type manifestEntry struct {
	File    string   `json:"file"`
	Src     string   `json:"src"`
	IsEntry bool     `json:"isEntry"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
}

// Bundle is the resolved asset URL list for one entry. The caller renders these
// into <script>/<link>/<link rel=modulepreload> tags however it likes.
type Bundle struct {
	JS       []string
	CSS      []string
	Preloads []string
}

// parseManifest reads and decodes <distDir>/.vite/manifest.json from fsys.
func parseManifest(fsys fs.FS, distDir string) (map[string]manifestEntry, error) {
	name := path.Join(distDir, ".vite", "manifest.json")
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("vite: read manifest %s: %w", name, err)
	}
	var m map[string]manifestEntry
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("vite: parse manifest %s: %w", name, err)
	}
	return m, nil
}

// resolve walks the manifest for one entry, collecting the entry's JS file, the
// CSS of the entry and all transitively imported chunks (de-duplicated in
// encounter order), and the imported chunk files as module preloads. All URLs
// are prefixed with staticURL. Pure over the parsed manifest; cycle-safe.
func resolve(manifest map[string]manifestEntry, name, staticURL string) Bundle {
	entry, ok := manifest[name]
	if !ok {
		return Bundle{}
	}
	var b Bundle
	b.JS = []string{staticURL + entry.File}

	cssSeen := map[string]bool{}
	addCSS := func(files []string) {
		for _, f := range files {
			if !cssSeen[f] {
				cssSeen[f] = true
				b.CSS = append(b.CSS, staticURL+f)
			}
		}
	}
	addCSS(entry.CSS)

	visited := map[string]bool{name: true}
	var walk func(keys []string)
	walk = func(keys []string) {
		for _, k := range keys {
			if visited[k] {
				continue
			}
			visited[k] = true
			chunk, ok := manifest[k]
			if !ok {
				continue
			}
			b.Preloads = append(b.Preloads, staticURL+chunk.File)
			addCSS(chunk.CSS)
			walk(chunk.Imports)
		}
	}
	walk(entry.Imports)
	return b
}
