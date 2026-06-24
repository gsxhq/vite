package vite

import (
	"fmt"
	"io/fs"
)

// Config configures a Vite integration. The zero DevURL selects prod mode.
type Config struct {
	DevURL    string // running Vite dev server origin; "" → prod
	DevBase   string // dev base path (vite.config base); default "/"
	Dist      fs.FS  // embedded prod build output (holds .vite/manifest.json); required in prod
	DistDir   string // subpath within Dist for manifest+assets; default "."
	StaticURL string // URL prefix prod assets serve under; default "/static/"
}

// Vite resolves Vite entries to asset URLs. Safe for concurrent use; build once
// at startup and share across requests.
type Vite struct {
	dev       bool
	devBase   string
	staticURL string
	distDir   string
	dist      fs.FS
	manifest  map[string]manifestEntry
}

// New builds a *Vite. In prod (DevURL == "") it reads and parses the manifest
// from Dist and returns an error if Dist is nil or the manifest is missing or
// invalid. In dev it performs no I/O.
func New(cfg Config) (*Vite, error) {
	v := &Vite{
		dev:       cfg.DevURL != "",
		devBase:   orDefault(cfg.DevBase, "/"),
		staticURL: orDefault(cfg.StaticURL, "/static/"),
		distDir:   orDefault(cfg.DistDir, "."),
		dist:      cfg.Dist,
	}
	if !v.dev {
		if cfg.Dist == nil {
			return nil, fmt.Errorf("vite: prod mode (empty DevURL) requires Config.Dist")
		}
		m, err := parseManifest(cfg.Dist, v.distDir)
		if err != nil {
			return nil, err
		}
		v.manifest = m
	}
	return v, nil
}

// Dev reports whether the integration is in dev mode.
func (v *Vite) Dev() bool { return v.dev }

// Entry resolves one Vite entry (the manifest key / source path, e.g.
// "web/main.js") to its asset URLs. Never panics; an unknown prod entry yields
// an empty Bundle.
func (v *Vite) Entry(name string) Bundle {
	if v.dev {
		return Bundle{JS: []string{v.devBase + "@vite/client", v.devBase + name}}
	}
	return resolve(v.manifest, name, v.staticURL)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
