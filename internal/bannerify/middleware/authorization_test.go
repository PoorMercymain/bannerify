package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/PoorMercymain/bannerify/internal/bannerify/domain/mocks"
	"github.com/PoorMercymain/bannerify/internal/bannerify/handlers"
	"github.com/PoorMercymain/bannerify/internal/bannerify/service"
	"github.com/PoorMercymain/bannerify/pkg/jwt"
)

func testRouter(t *testing.T) *http.ServeMux {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mux := http.NewServeMux()

	aur := mocks.NewMockAuthorizationRepository(ctrl)
	aus := service.NewAuthorization(aur)
	auh := handlers.NewAuthorization(aus, "")

	mux.Handle("GET /admin", AdminRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), auh.JWTKey))
	mux.Handle("GET /user", AuthorizationRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), auh.JWTKey))

	return mux
}

func request(t *testing.T, ts *httptest.Server, code int, method string, content string, body string, endpoint string, authorization string) *http.Response {
	req, err := http.NewRequest(method, ts.URL+endpoint, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", content)
	if authorization != "" {
		req.Header.Set("token", authorization)
	}

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, code, resp.StatusCode)

	return resp
}

func TestAdminRequired(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))

	defer ts.Close()

	tokenStrNoAdmin, err := jwt.CreateJWT(false, []byte(""), time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	tokenStrAdmin, err := jwt.CreateJWT(true, []byte(""), time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	wrongToken, err := jwt.CreateJWT(true, []byte("abcd"), time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	var testTable = []struct {
		endpoint      string
		method        string
		content       string
		code          int
		body          string
		authorization string
	}{
		{
			"/admin",
			http.MethodGet,
			"",
			http.StatusUnauthorized,
			"",
			"",
		},
		{
			"/admin",
			http.MethodGet,
			"",
			http.StatusForbidden,
			"",
			tokenStrNoAdmin,
		},
		{
			"/admin",
			http.MethodGet,
			"",
			http.StatusUnauthorized,
			"",
			wrongToken,
		},
		{
			"/admin",
			http.MethodGet,
			"",
			http.StatusOK,
			"",
			tokenStrAdmin,
		},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.content, testCase.body, testCase.endpoint, testCase.authorization)
		resp.Body.Close()
	}
}

func TestAuthorizationRequired(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))

	defer ts.Close()

	tokenStrNoAdmin, err := jwt.CreateJWT(false, []byte(""), time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	tokenStrAdmin, err := jwt.CreateJWT(true, []byte(""), time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	wrongToken, err := jwt.CreateJWT(true, []byte("abcd"), time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	var testTable = []struct {
		endpoint      string
		method        string
		content       string
		code          int
		body          string
		authorization string
	}{
		{
			"/user",
			http.MethodGet,
			"",
			http.StatusUnauthorized,
			"",
			"",
		},
		{
			"/user",
			http.MethodGet,
			"",
			http.StatusUnauthorized,
			"",
			wrongToken,
		},
		{
			"/user",
			http.MethodGet,
			"",
			http.StatusOK,
			"",
			tokenStrNoAdmin,
		},
		{
			"/user",
			http.MethodGet,
			"",
			http.StatusOK,
			"",
			tokenStrAdmin,
		},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.content, testCase.body, testCase.endpoint, testCase.authorization)
		resp.Body.Close()
	}
}
