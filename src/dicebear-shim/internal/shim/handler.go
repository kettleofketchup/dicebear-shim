package shim

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	DiceBearURL  string
	DefaultStyle string
	DefaultSize  string
	CacheMaxAge  int
}

type Handler struct {
	cfg    Config
	client *http.Client
	log    *slog.Logger
}

func NewHandler(cfg Config, log *slog.Logger) *Handler {
	return &Handler{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// ServeHTTP handles requests to /avatar/{hash}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract hash from path: /avatar/{hash} or /avatar/{hash}.png etc
	path := strings.TrimPrefix(r.URL.Path, "/avatar/")
	if path == "" || path == r.URL.Path {
		http.NotFound(w, r)
		return
	}

	// Strip file extensions (.jpg, .png, .gif, .webp)
	hash := path
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp"} {
		hash = strings.TrimSuffix(hash, ext)
	}

	// Parse Gravatar query params
	q := r.URL.Query()

	size := firstNonEmpty(q.Get("s"), q.Get("size"), h.cfg.DefaultSize)
	gravatarDefault := firstNonEmpty(q.Get("d"), q.Get("default"))
	style := ResolveDiceBearStyle(gravatarDefault, h.cfg.DefaultStyle)

	// Build DICEbear URL
	target := fmt.Sprintf("%s/9.x/%s/png?seed=%s&size=%s",
		strings.TrimRight(h.cfg.DiceBearURL, "/"),
		url.PathEscape(style),
		url.QueryEscape(hash),
		url.QueryEscape(size),
	)

	h.log.Info("proxying avatar request",
		"hash", hash,
		"style", style,
		"size", size,
		"target", target,
	)

	// Proxy to DICEbear
	resp, err := h.client.Get(target)
	if err != nil {
		h.log.Error("backend request failed", "error", err)
		http.Error(w, "avatar backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", h.cfg.CacheMaxAge))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
