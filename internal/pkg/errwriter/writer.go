package errwriter

import (
	"encoding/json"
	"net/http"

	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
	"github.com/PoorMercymain/bannerify/pkg/logger"
)

func WriteHTTPError(w http.ResponseWriter, err error, statusCode int, prefix string) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(domain.JSONError{Err: err.Error()})
	if err != nil {
		logger.Logger().Errorln(prefix, err.Error())
	}
}
