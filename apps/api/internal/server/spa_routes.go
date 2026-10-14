package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// spaRoutes knows which paths the web app's router has a page for, from the
// dist/spa-routes.json the prerender step writes, so handleSPA can answer any
// other path with a real 404 (dist/404.html) instead of the app shell with
// 200 — a "soft 404" that crawlers index and uptime monitors read as up.
// Loaded once, on first use: dist doesn't change while the server runs.
type spaRoutes struct {
	once     sync.Once
	patterns [][]string // nil when there's no spa-routes.json: every path is the app's
	notFound []byte
}

// notFoundPage returns the page to send with a 404 for path, or nil when
// path is one of the app's routes.
func (s *spaRoutes) notFoundPage(staticDir, path string) []byte {
	s.once.Do(func() { s.load(staticDir) })
	if s.patterns == nil {
		return nil
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	for _, p := range s.patterns {
		if matchRoute(p, segments) {
			return nil
		}
	}
	return s.notFound
}

func (s *spaRoutes) load(staticDir string) {
	raw, err := os.ReadFile(filepath.Join(staticDir, "spa-routes.json"))
	if err != nil {
		return
	}
	var paths []string
	if json.Unmarshal(raw, &paths) != nil {
		return
	}
	page, err := os.ReadFile(filepath.Join(staticDir, "404.html"))
	if err != nil {
		if page, err = os.ReadFile(filepath.Join(staticDir, "index.html")); err != nil {
			return
		}
	}
	for _, p := range paths {
		s.patterns = append(s.patterns, strings.Split(strings.Trim(p, "/"), "/"))
	}
	s.notFound = page
}

// matchRoute reports whether a path's segments fit a route pattern's, where
// a ":param" segment matches any one non-empty segment.
func matchRoute(pattern, segments []string) bool {
	if len(pattern) != len(segments) {
		return false
	}
	for i, p := range pattern {
		if strings.HasPrefix(p, ":") {
			if segments[i] == "" {
				return false
			}
			continue
		}
		if p != segments[i] {
			return false
		}
	}
	return true
}
