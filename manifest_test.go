package vite

import (
	"slices"
	"testing"
	"testing/fstest"
)

func TestResolveProd(t *testing.T) {
	m := map[string]manifestEntry{
		"web/main.js": {
			File:    "assets/main-AAA.js",
			Src:     "web/main.js",
			IsEntry: true,
			CSS:     []string{"assets/main-CSS.css"},
			Imports: []string{"_shared-BBB.js"},
		},
		"_shared-BBB.js": {
			File:    "assets/shared-BBB.js",
			CSS:     []string{"assets/shared-CSS.css"},
			Imports: []string{"_dep-CCC.js"},
		},
		"_dep-CCC.js": {
			File: "assets/dep-CCC.js",
			CSS:  []string{"assets/main-CSS.css"}, // duplicate of entry CSS
		},
	}
	b := resolve(m, "web/main.js", "/static/")
	if !slices.Equal(b.JS, []string{"/static/assets/main-AAA.js"}) {
		t.Fatalf("JS = %v", b.JS)
	}
	// entry css + shared css + dep css (dup removed), deduped, prefixed:
	if !slices.Equal(b.CSS, []string{"/static/assets/main-CSS.css", "/static/assets/shared-CSS.css"}) {
		t.Fatalf("CSS = %v", b.CSS)
	}
	// transitively imported chunk files as preloads, prefixed:
	if !slices.Equal(b.Preloads, []string{"/static/assets/shared-BBB.js", "/static/assets/dep-CCC.js"}) {
		t.Fatalf("Preloads = %v", b.Preloads)
	}
}

func TestResolveUnknownEntry(t *testing.T) {
	b := resolve(map[string]manifestEntry{}, "nope", "/static/")
	if len(b.JS) != 0 || len(b.CSS) != 0 || len(b.Preloads) != 0 {
		t.Fatalf("expected empty bundle, got %+v", b)
	}
}

func TestResolveCycleTerminates(t *testing.T) {
	m := map[string]manifestEntry{
		"a.js": {File: "a.js", Imports: []string{"b.js"}},
		"b.js": {File: "b.js", Imports: []string{"a.js"}}, // cycle back to entry
	}
	b := resolve(m, "a.js", "/static/")
	if !slices.Equal(b.Preloads, []string{"/static/b.js"}) {
		t.Fatalf("Preloads = %v", b.Preloads)
	}
}

func TestParseManifest(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/.vite/manifest.json": &fstest.MapFile{
			Data: []byte(`{"web/main.js":{"file":"assets/main-AAA.js","isEntry":true}}`),
		},
	}
	m, err := parseManifest(fsys, "dist")
	if err != nil {
		t.Fatal(err)
	}
	if m["web/main.js"].File != "assets/main-AAA.js" {
		t.Fatalf("got %+v", m)
	}
}

func TestParseManifestMissing(t *testing.T) {
	if _, err := parseManifest(fstest.MapFS{}, "dist"); err == nil {
		t.Fatal("expected error for missing manifest")
	}
}
