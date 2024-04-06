package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	appErrors "github.com/PoorMercymain/bannerify/errors"
	"github.com/PoorMercymain/bannerify/internal/bannerify/domain"
	"github.com/PoorMercymain/bannerify/internal/pkg/errwriter"
	"github.com/PoorMercymain/bannerify/pkg/jwt"
	"github.com/PoorMercymain/bannerify/pkg/logger"
	"github.com/PoorMercymain/bannerify/pkg/reqval"
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

type authorization struct {
	srv    domain.AuthorizationService
	JWTKey string
}

func NewAuthorization(srv domain.AuthorizationService, jwtKey string) *authorization {
	return &authorization{srv: srv, JWTKey: jwtKey}
}

func (h *authorization) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.Register:"

	err := reqval.ValidateJSONRequest(r)
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logErrPrefix)
		return
	}

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var authData domain.AuthorizationData
	if err = d.Decode(&authData); err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logErrPrefix)
		return
	}

	var isAdmin bool
	if r.Header.Get("admin") == "true" {
		isAdmin = true
	}

	err = h.srv.Register(r.Context(), authData.Login, authData.Password, isAdmin)
	if err != nil {
		if errors.Is(err, appErrors.ErrAlreadyRegistered) {
			errwriter.WriteHTTPError(w, appErrors.ErrAlreadyRegistered, http.StatusConflict, logErrPrefix)
			return
		}

		logger.Logger().Errorln(logErrPrefix, err)
		errwriter.WriteHTTPError(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError, logErrPrefix)
		return
	}

	tokenStr, err := jwt.CreateJWT(isAdmin, []byte(h.JWTKey), time.Now().Add(24*time.Hour))
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err)
		errwriter.WriteHTTPError(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError, logErrPrefix)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(domain.Token{Token: tokenStr})
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err)
	}
}

func (h *authorization) LogIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.LogIn:"

	err := reqval.ValidateJSONRequest(r)
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logErrPrefix)
		return
	}

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var authData domain.AuthorizationData
	if err = d.Decode(&authData); err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logErrPrefix)
		return
	}

	err = h.srv.CheckAuth(r.Context(), authData.Login, authData.Password)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			errwriter.WriteHTTPError(w, appErrors.ErrUserNotFound, http.StatusNotFound, logErrPrefix)
			return
		}

		if errors.Is(err, appErrors.ErrWrongPassword) {
			errwriter.WriteHTTPError(w, appErrors.ErrWrongPassword, http.StatusUnauthorized, logErrPrefix)
			return
		}

		logger.Logger().Errorln(logErrPrefix, err)
		errwriter.WriteHTTPError(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError, logErrPrefix)
		return
	}

	isAdmin, err := h.srv.IsAdmin(r.Context(), authData.Login)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			errwriter.WriteHTTPError(w, appErrors.ErrUserNotFound, http.StatusNotFound, logErrPrefix)
			return
		}

		logger.Logger().Errorln(logErrPrefix, err)
		errwriter.WriteHTTPError(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError, logErrPrefix)
		return
	}

	logger.Logger().Infoln(isAdmin)
	tokenStr, err := jwt.CreateJWT(isAdmin, []byte(h.JWTKey), time.Now().Add(24*time.Hour))
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err)
		errwriter.WriteHTTPError(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError, logErrPrefix)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	e := json.NewEncoder(w)
	err = e.Encode(domain.Token{Token: tokenStr})
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err)
	}
}
