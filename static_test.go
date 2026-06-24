package vite

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestStaticHandlerServesAssets(t *testing.T) {
	fsys := fstest.MapFS{
		"dist/.vite/manifest.json": &fstest.MapFile{Data: []byte(`{}`)},
		"dist/assets/main-AAA.js":  &fstest.MapFile{Data: []byte("console.log(1)")},
	}
	v, err := New(Config{Dist: fsys, DistDir: "dist"})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(v.StaticHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/static/assets/main-AAA.js")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "console.log(1)" {
		t.Fatalf("body = %q", body)
	}
}
