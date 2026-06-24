package vite

import (
	"context"
	"net/http"
	"time"
)

// NotifyReload POSTs to <devURL>/__reload so a Vite plugin exposing that endpoint
// broadcasts a browser full-reload. Call it once after the HTTP server's
// listeners are up. Dev-only: a "" devURL is a no-op. Runs in a goroutine with a
// brief retry loop (covers the cold-start race where the Go server beats Vite to
// the port).
func NotifyReload(devURL string) {
	if devURL == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		for range 10 {
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, devURL+"/__reload", nil)
			if err != nil {
				return
			}
			if resp, err := http.DefaultClient.Do(req); err == nil {
				resp.Body.Close()
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(150 * time.Millisecond):
			}
		}
	}()
}
