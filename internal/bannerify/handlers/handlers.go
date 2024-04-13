package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"
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

func (h *banner) GetBanner(isAdmin bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		var dbRequired bool
		if r.Header.Get("use_last_revision") == "true" {
			dbRequired = true
		} else if r.Header.Get("use_last_revision") == "false" {
			dbRequired = false
		} else if r.Header.Get("use_last_revision") != "" {
			errwriter.WriteHTTPError(w, appErrors.ErrUseLastRevisionNotBool, http.StatusBadRequest, logErrPrefix)
			return
		}

		banner, err := h.srv.GetBanner(r.Context(), tagID, featureID, isAdmin, dbRequired)
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

	var featureID *int
	var tagID *int

	if featureIDStr != "" {
		featureIDBuf, err := strconv.Atoi(featureIDStr)
		if err != nil {
			errwriter.WriteHTTPError(w, appErrors.ErrFeatureIsNotANumber, http.StatusBadRequest, logErrPrefix)
			return
		}

		featureID = &featureIDBuf
	}

	if tagIDStr != "" {
		tagIDBuf, err := strconv.Atoi(tagIDStr)
		if err != nil {
			errwriter.WriteHTTPError(w, appErrors.ErrTagIsNotANumber, http.StatusBadRequest, logErrPrefix)
			return
		}

		tagID = &tagIDBuf
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

func (h *banner) ListVersions(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.ListVersions:"

	bannerIDStr := r.PathValue("banner_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	if limitStr == "" {
		limitStr = "3"
	}

	if offsetStr == "" {
		offsetStr = "0"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrLimitIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrOffsetIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	if bannerIDStr == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrNoBannerIDProvided, http.StatusBadRequest, logErrPrefix)
		return
	}

	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrBannerIDIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	if limit < 0 || limit > 100 {
		errwriter.WriteHTTPError(w, appErrors.ErrLimitNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	if offset < 0 {
		errwriter.WriteHTTPError(w, appErrors.ErrOffsetNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	versions, err := h.srv.ListVersions(r.Context(), bannerID, limit, offset)
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logErrPrefix)
		return
	}

	if len(versions) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(versions)
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err)
	}
}

func (h *banner) ChooseVersion(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.ChooseVersion:"

	bannerIDStr := r.PathValue("banner_id")
	versionIDStr := r.URL.Query().Get("version_id")

	if bannerIDStr == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrNoBannerIDProvided, http.StatusBadRequest, logErrPrefix)
		return
	}

	if versionIDStr == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrNoVersionIDProvided, http.StatusBadRequest, logErrPrefix)
		return
	}

	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrBannerIDIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrVersionIDIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	if bannerID < 1 {
		errwriter.WriteHTTPError(w, appErrors.ErrBannerIDNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	if versionID < 1 {
		errwriter.WriteHTTPError(w, appErrors.ErrVersionIDNotInRange, http.StatusBadRequest, logErrPrefix)
		return
	}

	err = h.srv.ChooseVersion(r.Context(), bannerID, versionID)
	if err != nil {
		if errors.Is(err, appErrors.ErrBannerTagUniqueViolation) {
			errwriter.WriteHTTPError(w, appErrors.ErrBannerTagUniqueViolation, http.StatusConflict, logErrPrefix)
			return
		}

		if errors.Is(err, appErrors.ErrBannerNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, appErrors.ErrVersionNotFound) {
			errwriter.WriteHTTPError(w, appErrors.ErrVersionNotFound, http.StatusBadRequest, logErrPrefix)
			return
		}

		logger.Logger().Errorln(logErrPrefix, err.Error())
		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logErrPrefix)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *banner) CreateBanner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.CreateBanner:"

	err := reqval.ValidateJSONRequest(r)
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logErrPrefix)
		return
	}

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var banner domain.Banner
	if err = d.Decode(&banner); err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logErrPrefix)
		return
	}

	if banner.Content == nil || banner.FeatureID == nil || banner.IsActive == nil || banner.TagIDs == nil {
		errwriter.WriteHTTPError(w, appErrors.ErrBannerFieldNotProvided, http.StatusBadRequest, logErrPrefix)
		return
	}

	bannerID, err := h.srv.CreateBanner(r.Context(), banner)
	if err != nil {
		if errors.Is(err, appErrors.ErrBannerTagUniqueViolation) {
			errwriter.WriteHTTPError(w, appErrors.ErrBannerTagUniqueViolation, http.StatusBadRequest, logErrPrefix)
			return
		}

		logger.Logger().Errorln(logErrPrefix, err.Error())
		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logErrPrefix)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err = json.NewEncoder(w).Encode(domain.BannerID{ID: bannerID}); err != nil {
		logger.Logger().Errorln(logErrPrefix, err)
	}
}

func (h *banner) UpdateBanner(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.UpdateBanner:"

	bannerIDStr := r.PathValue("id")

	if bannerIDStr == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrNoBannerIDProvided, http.StatusBadRequest, logErrPrefix)
		return
	}

	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrBannerIDIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	err = reqval.ValidateJSONRequest(r)
	if err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logErrPrefix)
		return
	}

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	var banner domain.Banner
	if err = d.Decode(&banner); err != nil {
		errwriter.WriteHTTPError(w, err, http.StatusBadRequest, logErrPrefix)
		return
	}

	if banner.Content == nil && banner.FeatureID == nil && banner.IsActive == nil && banner.TagIDs == nil {
		errwriter.WriteHTTPError(w, appErrors.ErrNoBannerFieldsProvided, http.StatusBadRequest, logErrPrefix)
		return
	}

	err = h.srv.UpdateBanner(r.Context(), bannerID, banner)
	if err != nil {
		logger.Logger().Errorln(logErrPrefix, err.Error())
		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logErrPrefix)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *banner) DeleteBannerByID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	const logErrPrefix = "handlers.DeleteBannerByID:"

	bannerIDStr := r.PathValue("id")

	if bannerIDStr == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrNoBannerIDProvided, http.StatusBadRequest, logErrPrefix)
		return
	}

	bannerID, err := strconv.Atoi(bannerIDStr)
	if err != nil {
		errwriter.WriteHTTPError(w, appErrors.ErrBannerIDIsNotANumber, http.StatusBadRequest, logErrPrefix)
		return
	}

	err = h.srv.DeleteBannerByID(r.Context(), bannerID)
	if err != nil {
		if errors.Is(err, appErrors.ErrBannerNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		logger.Logger().Errorln(logErrPrefix, err.Error())
		errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logErrPrefix)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *banner) DeleteBannerByTagOrFeature(deleteCtx context.Context, wg *sync.WaitGroup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		const logErrPrefix = "handlers.DeleteBannerByTagOrFeature:"

		tagIDStr := r.URL.Query().Get("tag_id")
		featureIDStr := r.URL.Query().Get("feature_id")

		var tagID *int
		if tagIDStr != "" {
			tagIDbuf, err := strconv.Atoi(tagIDStr)
			if err != nil {
				errwriter.WriteHTTPError(w, appErrors.ErrTagIsNotANumber, http.StatusBadRequest, logErrPrefix)
				return
			}

			tagID = &tagIDbuf
		}

		var featureID *int
		if featureIDStr != "" {
			featureIDbuf, err := strconv.Atoi(featureIDStr)
			if err != nil {
				errwriter.WriteHTTPError(w, appErrors.ErrFeatureIsNotANumber, http.StatusBadRequest, logErrPrefix)
				return
			}

			featureID = &featureIDbuf
		}

		if tagID == nil && featureID == nil {
			errwriter.WriteHTTPError(w, appErrors.ErrTagOrFeatureNotProvided, http.StatusBadRequest, logErrPrefix)
			return
		}

		err := h.srv.DeleteBannerByTagOrFeature(r.Context(), deleteCtx, tagID, featureID)
		if err != nil {
			if errors.Is(err, appErrors.ErrBannerNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			logger.Logger().Errorln(logErrPrefix, err.Error())
			errwriter.WriteHTTPError(w, err, http.StatusInternalServerError, logErrPrefix)
			return
		}

		w.WriteHeader(http.StatusAccepted)
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
	} else if r.Header.Get("admin") != "false" && r.Header.Get("admin") != "" {
		errwriter.WriteHTTPError(w, appErrors.ErrWrongAdminHeader, http.StatusBadRequest, logErrPrefix)
		return
	}

	if authData.Login == "" || authData.Password == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrNoLoginOrPassword, http.StatusBadRequest, logErrPrefix)
		return
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

	if authData.Login == "" || authData.Password == "" {
		errwriter.WriteHTTPError(w, appErrors.ErrNoLoginOrPassword, http.StatusBadRequest, logErrPrefix)
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
