package vite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromContextRoundTrip(t *testing.T) {
	v := &Vite{} // identity is all that matters here
	ctx := NewContext(context.Background(), v)
	if got := FromContext(ctx); got != v {
		t.Fatalf("FromContext = %p, want %p", got, v)
	}
}

func TestFromContextAbsent(t *testing.T) {
	if got := FromContext(context.Background()); got != nil {
		t.Fatalf("FromContext on empty ctx = %v, want nil", got)
	}
}

func TestMiddlewareInjects(t *testing.T) {
	v, err := New(Config{DevURL: "http://localhost:5173"})
	if err != nil {
		t.Fatal(err)
	}
	var seen *Vite
	h := v.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if seen != v {
		t.Fatalf("middleware did not inject *Vite into request context (seen=%p, want=%p)", seen, v)
	}
}
