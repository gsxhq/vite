# vite

Generic Go ↔ Vite integration: manifest-driven prod asset URLs, live dev-server
URLs in development, a static file handler for embedded assets, and a reload
notifier — stdlib-only, no framework dependency.

## Install

```
go get github.com/gsxhq/vite
```

## Usage

```go
import (
    "embed"
    "log"
    "net/http"
    "os"

    "github.com/gsxhq/vite"
)

//go:embed all:dist
var distFS embed.FS

func main() {
    devURL := os.Getenv("VITE_DEV_URL") // "" in prod
    v, err := vite.New(vite.Config{
        DevURL:  devURL,
        Dist:    distFS,
        DistDir: "dist", // //go:embed all:dist nests under dist/
    })
    if err != nil {
        log.Fatal(err)
    }
    mux := http.NewServeMux()
    if !v.Dev() {
        mux.Handle("/static/", v.StaticHandler())
    }
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        b := v.Entry("web/main.js")
        // render b.CSS as <link rel=stylesheet>, b.JS as <script type=module>,
        // b.Preloads as <link rel=modulepreload> — in your template of choice.
        _ = b
    })
    vite.NotifyReload(devURL) // dev-only; no-op when devURL == ""
    http.ListenAndServe(":7777", mux)
}
```

One `DevURL` environment variable flips between dev and prod: set it to the
running Vite dev server origin (e.g. `http://localhost:5173`) during development
and leave it empty for production.

## Dev vs prod

`Entry` behaves differently depending on mode:

| | Dev (`DevURL != ""`) | Prod (`DevURL == ""`) |
|---|---|---|
| **JS** | `["/<base>@vite/client", "/<base><entry>"]` — default `DevBase "/"`: `["/@vite/client", "/web/main.js"]` | Hashed JS file from the manifest, prefixed with `StaticURL` |
| **CSS** | `[]` (Vite HMR injects styles) | Hashed CSS files from the manifest, prefixed with `StaticURL` |
| **Preloads** | `[]` | Hashed JS chunks for `<link rel=modulepreload>`, prefixed with `StaticURL` |

In dev the browser talks directly to the Vite dev server; the Go server just
injects the two script tags. In prod all assets are resolved from the embedded
`dist/.vite/manifest.json` and served by `StaticHandler`.

## `//go:embed all:dist` and build order

When you embed the `dist/` directory with `//go:embed all:dist`, Go nests the
files under a `dist/` path inside the `embed.FS`. Pass `DistDir: "dist"` so the
library strips that prefix when reading the manifest and serving assets.

**Prod build order:**

1. `vite build` — writes `dist/` (including `.vite/manifest.json`).
2. `go build` — embeds `dist/` into the binary.

Running `go build` before `vite build` will fail or embed an empty/stale
`dist/`.

## API reference

### `Config`

```go
type Config struct {
    DevURL    string // running Vite dev server origin; "" → prod
    DevBase   string // dev base path (vite.config base); default "/"
    Dist      fs.FS  // embedded prod build output; required in prod
    DistDir   string // subpath within Dist for manifest+assets; default "."
    StaticURL string // URL prefix prod assets serve under; default "/static/"
}
```

| Field | Default | Notes |
|---|---|---|
| `DevURL` | `""` | Empty → prod mode. Set to Vite dev server origin, e.g. `http://localhost:5173`. |
| `DevBase` | `"/"` | Matches `base` in `vite.config`. Prefixed onto `@vite/client` and the entry in dev. |
| `Dist` | — | Required in prod. Pass your `embed.FS`. Ignored in dev. |
| `DistDir` | `"."` | Subpath within `Dist` where manifest and assets live. Use `"dist"` with `//go:embed all:dist`. |
| `StaticURL` | `"/static/"` | URL prefix under which prod assets are served. |

### `func New(cfg Config) (*Vite, error)`

Builds and returns a `*Vite`. In prod (`DevURL == ""`), reads and parses
`<DistDir>/.vite/manifest.json` from `Dist`; returns an error if `Dist` is nil
or the manifest is missing or invalid. In dev it performs no I/O. Safe to build
once at startup and share across goroutines.

### `func (v *Vite) Dev() bool`

Reports whether the integration is in dev mode (`DevURL != ""`).

### `func (v *Vite) Entry(name string) Bundle`

Resolves one Vite entry (the manifest key / source path, e.g. `"web/main.js"`)
to its asset URLs. Never panics; an unknown prod entry yields an empty `Bundle`.

```go
type Bundle struct {
    JS       []string // script URLs (<script type=module>)
    CSS      []string // stylesheet URLs (<link rel=stylesheet>)
    Preloads []string // chunk URLs (<link rel=modulepreload>)
}
```

### `func (v *Vite) StaticHandler() http.Handler`

Returns an `http.Handler` that serves the embedded prod assets (from `Config.Dist`
rooted at `DistDir`) under `Config.StaticURL`. Mount it only in prod (`if !v.Dev()`)
or always — the `/static/` path is never hit in dev.

### `func NotifyReload(devURL string)`

POSTs to `<devURL>/__reload` so a Vite plugin exposing that endpoint broadcasts a
browser full-reload. Call it once after the HTTP server's listeners are up. A
`""` devURL is a no-op. Internally spawns a goroutine with a brief retry loop to
cover the cold-start race where the Go server beats Vite to the port.

### Context helpers (v0.2.0)

For passing the `*Vite` through the request context instead of as a parameter to
every handler/template:

```go
func NewContext(ctx context.Context, v *Vite) context.Context // stash *Vite
func FromContext(ctx context.Context) *Vite                   // retrieve (nil if absent)
func (v *Vite) Middleware(next http.Handler) http.Handler     // inject per request
```

Wrap your handler with `v.Middleware(mux)`, then read the bundle anywhere with
`vite.FromContext(ctx).Entry("web/main.js")`.

## License

MIT
