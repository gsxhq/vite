package vite

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNotifyReloadPosts(t *testing.T) {
	got := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- r.Method + " " + r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	NotifyReload(srv.URL)
	select {
	case g := <-got:
		if g != "POST /__reload" {
			t.Fatalf("got %q, want POST /__reload", g)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("NotifyReload did not POST /__reload")
	}
}

func TestNotifyReloadEmptyNoop(t *testing.T) {
	got := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got <- struct{}{}
	}))
	defer srv.Close()

	NotifyReload("") // must not POST anywhere
	select {
	case <-got:
		t.Fatal(`NotifyReload("") should not POST`)
	case <-time.After(200 * time.Millisecond):
		// success: nothing received
	}
}
