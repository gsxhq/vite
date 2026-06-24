package vite

import (
	"slices"
	"testing"
	"testing/fstest"
)

func TestEntryDev(t *testing.T) {
	v, err := New(Config{DevURL: "http://localhost:5173"})
	if err != nil {
		t.Fatal(err)
	}
	if !v.Dev() {
		t.Fatal("expected dev mode")
	}
	b := v.Entry("web/main.js")
	if !slices.Equal(b.JS, []string{"/@vite/client", "/web/main.js"}) {
		t.Fatalf("JS = %v", b.JS)
	}
	if len(b.CSS) != 0 {
		t.Fatalf("dev CSS should be empty, got %v", b.CSS)
	}
}

func TestEntryDevCustomBase(t *testing.T) {
	v, err := New(Config{DevURL: "http://x", DevBase: "/__vite/"})
	if err != nil {
		t.Fatal(err)
	}
	b := v.Entry("web/main.js")
	if !slices.Equal(b.JS, []string{"/__vite/@vite/client", "/__vite/web/main.js"}) {
		t.Fatalf("JS = %v", b.JS)
	}
}

func TestNewProdCustomStaticURL(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/.vite/manifest.json": &fstest.MapFile{
			Data: []byte(`{"web/main.js":{"file":"assets/main-BBB.js","css":["assets/main-CSS.css"]}}`),
		},
	}
	v, err := New(Config{Dist: fsys, DistDir: "dist", StaticURL: "/assets/"})
	if err != nil {
		t.Fatal(err)
	}
	b := v.Entry("web/main.js")
	if !slices.Equal(b.JS, []string{"/assets/assets/main-BBB.js"}) {
		t.Fatalf("JS = %v", b.JS)
	}
	if !slices.Equal(b.CSS, []string{"/assets/assets/main-CSS.css"}) {
		t.Fatalf("CSS = %v", b.CSS)
	}
}

func TestNewProdParsesManifestAndResolves(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/.vite/manifest.json": &fstest.MapFile{
			Data: []byte(`{"web/main.js":{"file":"assets/main-AAA.js","css":["assets/main-CSS.css"]}}`),
		},
	}
	v, err := New(Config{Dist: fsys, DistDir: "dist"})
	if err != nil {
		t.Fatal(err)
	}
	if v.Dev() {
		t.Fatal("expected prod mode")
	}
	b := v.Entry("web/main.js")
	if !slices.Equal(b.JS, []string{"/static/assets/main-AAA.js"}) {
		t.Fatalf("JS = %v", b.JS)
	}
	if !slices.Equal(b.CSS, []string{"/static/assets/main-CSS.css"}) {
		t.Fatalf("CSS = %v", b.CSS)
	}
}

func TestNewProdMissingManifestErrors(t *testing.T) {
	if _, err := New(Config{Dist: fstest.MapFS{}, DistDir: "dist"}); err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestNewProdNilDistErrors(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected error for prod mode without Dist")
	}
}
