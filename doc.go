// Package vite integrates a Vite-built frontend into a Go server: it resolves a
// Vite build manifest to hashed asset URLs in production, points at the Vite dev
// server in development, serves the embedded production assets, and notifies the
// Vite dev server to reload — all behind one dev boolean.
//
// It depends only on the standard library and on Vite's manifest format; it has
// no knowledge of any Go HTML templating library.
package vite
