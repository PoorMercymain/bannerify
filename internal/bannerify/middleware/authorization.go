package middleware

import (
	"net/http"

	appErrors "github.com/PoorMercymain/bannerify/errors"
	"github.com/PoorMercymain/bannerify/pkg/jwt"
)

func AdminRequired(next http.Handler, jwtKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, err := checkIsAdmin(r, jwtKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !isAdmin {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AuthorizationRequired(next http.Handler, jwtKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := checkIsAdmin(r, jwtKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ProvideIsAdmin(next func(bool) http.HandlerFunc, jwtKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, err := checkIsAdmin(r, jwtKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next(isAdmin).ServeHTTP(w, r)
	})
}

func checkIsAdmin(r *http.Request, jwtKey string) (bool, error) {
	authToken := r.Header.Get("token")
	if authToken == "" {
		return false, appErrors.ErrNoTokenProvided
	}

	isAdmin, err := jwt.CheckIsAdminInJWT(authToken, jwtKey)
	if err != nil {
		return false, appErrors.ErrTokenIsInvalid
	}

	return isAdmin, nil
}
