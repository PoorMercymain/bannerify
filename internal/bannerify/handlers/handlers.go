package handlers

import (
	"net/http"

	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
)

type banner struct {
	srv domain.BannerService
}

func NewBanner(srv domain.BannerService) *banner {
	return &banner{srv: srv}
}

func (h *banner) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.srv.Ping(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}