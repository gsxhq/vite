package vite

import (
	"context"
	"net/http"
)

// ctxKey is the unexported context key under which *Vite is stored.
type ctxKey struct{}

// NewContext returns a copy of ctx carrying v, retrievable with FromContext.
func NewContext(ctx context.Context, v *Vite) context.Context {
	return context.WithValue(ctx, ctxKey{}, v)
}

// FromContext returns the *Vite stored by NewContext, or nil if none is present.
// Templates and handlers use it to read the asset bundle from the request
// context instead of receiving it as a parameter.
func FromContext(ctx context.Context) *Vite {
	v, _ := ctx.Value(ctxKey{}).(*Vite)
	return v
}

// Middleware injects v into each request's context (via NewContext) before
// calling next, so downstream handlers and rendered templates can retrieve it
// with FromContext.
func (v *Vite) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), v)))
	})
}
