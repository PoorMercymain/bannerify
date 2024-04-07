package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

func (h *banner) GetBanner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.GetBanner:"

	tagIDStr := r.URL.Query().Get("tag_id")
	featureIDStr := r.URL.Query().Get("feature_id")

	if tagIDStr == "" || featureIDStr == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrTagOrFeatureNotProvided, http.StatusBadRequest, logErrPrefix)
		return
	}

	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrTagIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	featureID, err := strconv.Atoi(featureIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrFeatureIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	banner, err := h.srv.GetBanner(r.Context(), tagID, featureID)
	if err != nil {
		if errors.Is(err, appErrors.ErrBannerNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger().Errorln(logErrPrefix, err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(banner))
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err.Error())
	}
}

func (h *banner) ListBanners(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.ListBanners:"

	featureIDStr := r.URL.Query().Get("feature_id")
	tagIDStr := r.URL.Query().Get("tag_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if limitStr == "" {
		limitStr = "15"
	}

	if offsetStr == "" {
		offsetStr = "0"
	}

	if featureIDStr == "" {
		featureIDStr = "-1"
	} else if featureIDStr == "-1" {
		errwriter.WriteHTTPError(w, appErrors.ErrFeatureNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	if tagIDStr == "" {
		tagIDStr = "-1"
	} else if tagIDStr == "-1" {
		errwriter.WriteHTTPError(w, appErrors.ErrFeatureNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrLimitIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	if limit < 1 || limit > 100 {
		errwriter.WriteHTTPError(w, appErrors.ErrLimitNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrOffsetIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	if offset < 0 {
		errwriter.WriteHTTPError(w, appErrors.ErrOffsetNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	featureID, err := strconv.Atoi(featureIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrFeatureIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	if featureID < 1 && featureID != -1 {
		errwriter.WriteHTTPError(w, appErrors.ErrFeatureNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrTagIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	if tagID < 1 && tagID != -1 {
		errwriter.WriteHTTPError(w, appErrors.ErrTagNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	banners, err := h.srv.ListBanners(r.Context(), tagID, featureID, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Logger().Errorln(logErrPrefix, err.Error())
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if len(banners) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(banners)
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err)
	}
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenStr, err := jwt.CreateJWT(isAdmin, []byte(h.JWTKey), time.Now().Add(24*time.Hour))
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err)
		w.WriteHeader(http.StatusInternalServerError)
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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	isAdmin, err := h.srv.IsAdmin(r.Context(), authData.Login)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			errwriter.WriteHTTPError(w, appErrors.ErrUserNotFound, http.StatusNotFound, logErrPrefix)
			return
		}

		logger.Logger().Errorln(logErrPrefix, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Logger().Infoln(isAdmin)
	tokenStr, err := jwt.CreateJWT(isAdmin, []byte(h.JWTKey), time.Now().Add(24*time.Hour))
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err)

		w.WriteHeader(http.StatusInternalServerError)
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
