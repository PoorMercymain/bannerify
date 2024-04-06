package middleware

import (
	"net/http"

	"github.com/PoorMercymain/bannerify/pkg/jwt"
)

func AdminRequired(next http.Handler, jwtKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get("token")
		if authToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		isAdmin, err := jwt.CheckIsAdminInJWT(authToken, jwtKey)
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
		authToken := r.Header.Get("token")
		if authToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		_, err := jwt.CheckIsAdminInJWT(authToken, jwtKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
