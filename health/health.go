package health

import (
	"net/http"

	"github.com/taxfyle/go-httpkit/log"
)

type Handler struct {
	Path string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	logger.Debug("handling health request")

	route := r.URL.Path[len(h.Path):]
	logger.Debugf("handling route %s", route)

	switch route {
	case "/readiness":
		switch r.Method {
		case http.MethodGet:
			h.handleReadiness(w, r)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func (h *Handler) handleReadiness(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context())
	logger.Debug("handling readiness request")

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MatchPath(path string) bool {
	return len(path) >= len(h.Path) && path[:len(h.Path)] == h.Path
}
