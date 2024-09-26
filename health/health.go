package health

import (
	"net/http"

	"github.com/taxfyle/go-httpkit/v2/log"
)

type Handler struct {
}

func (h *Handler) GetReadiness(w http.ResponseWriter, r *http.Request) {
	logger := log.FromContext(r.Context()).Sugar()
	logger.Debug("handling health request")

	route := r.URL.Path
	logger.Debugf("handling route %s", route)

	logger.Debug("handling readiness request")

	w.WriteHeader(http.StatusNoContent)
}
