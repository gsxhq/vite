package vite

import (
	"io/fs"
	"net/http"
	"strings"
)

// StaticHandler serves the embedded prod assets (Config.Dist, rooted at DistDir)
// under Config.StaticURL. In dev it has no assets to serve and returns a
// not-found handler; mount it only in prod (or always — /static/ is never hit
// in dev). The StaticURL prefix is stripped without its trailing slash so the
// remaining request path keeps its leading slash for the file server.
func (v *Vite) StaticHandler() http.Handler {
	if v.dist == nil {
		return http.NotFoundHandler()
	}
	sub := v.dist
	if v.distDir != "." {
		if s, err := fs.Sub(v.dist, v.distDir); err == nil {
			sub = s
		}
	}
	prefix := strings.TrimSuffix(v.staticURL, "/")
	return http.StripPrefix(prefix, http.FileServerFS(sub))
}
