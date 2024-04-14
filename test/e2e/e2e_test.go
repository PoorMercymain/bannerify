//go:build e2e
package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/require"
)

type e2eConfig struct {
	ServicePort int    `env:"SERVICE_PORT" envDefault:"8080"`
	ServiceHost string `env:"SERVICE_HOST" envDefault:"bannerify-e2e"`
}

type auth struct {
	Token string `json:"token"`
}

type testTableElem struct {
	caseName string
	httpMethod string
	route string
	body string
	headers [][2]string
	expectedStatus int
	requireParsing bool
	parsedBody interface{}
}

type bannerListElement struct {
	BannerID  int             `json:"banner_id"`
	TagIDs    []int           `json:"tag_ids"`
	FeatureID int             `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

type versionListElement struct {
	VersionID int             `json:"version_id"`
	TagIDs    []int           `json:"tag_ids"`
	FeatureID int             `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	IsChosen  bool            `json:"is_chosen"`
}

type bannerID struct {
	ID int `json:"banner_id"`
}

func buildRequest(httpMethod string, route string, body string, headers [][2]string, cfg e2eConfig) (*http.Request, error) {
	req, err := http.NewRequest(httpMethod, fmt.Sprintf("http://%s:%d%s", cfg.ServiceHost, cfg.ServicePort, route), strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	for _, header := range headers {
		req.Header.Add(header[0], header[1])
	}

	return req, nil
}

func sendReq(t *testing.T, client *http.Client, req *http.Request, expectedStatus int, parsedBody interface{}, requireParsing bool) {
	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, expectedStatus, resp.StatusCode)

	if requireParsing {
		err = json.NewDecoder(resp.Body).Decode(parsedBody)
		require.NoError(t, err)
	}

	resp.Body.Close()
}

func TestPing(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodGet,
			route: "/ping",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin0\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user0\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "user token used",
			httpMethod: http.MethodGet,
			route: "/ping",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusForbidden,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "ping ok",
			httpMethod: http.MethodGet,
			route: "/ping",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "wrong method",
			httpMethod: http.MethodPost,
			route: "/ping",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusMethodNotAllowed,
			requireParsing: false,
			parsedBody: nil,
		},
	}

	var token string
	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "user token used" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
	}
}

func TestRegister(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var authData auth

	var testTable = []testTableElem {
		{
			caseName: "no Content-Type header",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user1\",\"password\": \"password\"}",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "no password",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user1\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "no login",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "empty json",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "duplicate login",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user1\", \"login\": \"user1\", \"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "user ok",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user1\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &authData,
		},
		{
			caseName: "register duplicate user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user1\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusConflict,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "admin ok",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin1\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &authData,
		},
		{
			caseName: "wrong admin header",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin2\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "123"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
	}

	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, testCase.headers, cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
	}
}

func TestLogin(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var authData auth

	var testTable = []testTableElem {
		{
			caseName: "no registration",
			httpMethod: http.MethodPost,
			route: "/acquire-token",
			body: "{\"login\": \"user3\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "no Content-Type header",
			httpMethod: http.MethodPost,
			route: "/acquire-token",
			body: "{\"login\": \"user3\",\"password\": \"password\"}",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "no login",
			httpMethod: http.MethodPost,
			route: "/acquire-token",
			body: "{\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "no password",
			httpMethod: http.MethodPost,
			route: "/acquire-token",
			body: "{\"login\": \"user3\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "empty json",
			httpMethod: http.MethodPost,
			route: "/acquire-token",
			body: "{}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "duplicate password",
			httpMethod: http.MethodPost,
			route: "/acquire-token",
			body: "{\"login\": \"user3\",\"password\": \"password\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "register ok",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user3\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: false,
			parsedBody: &authData,
		},
		{
			caseName: "login ok",
			httpMethod: http.MethodPost,
			route: "/acquire-token",
			body: "{\"login\": \"user3\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &authData,
		},
	}

	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, testCase.headers, cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
		if testCase.requireParsing {
			require.NotEmpty(t, len(authData.Token))
		}
	}
}

func TestGetBanner(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth
	var banner, updatedBanner, cachedBanner json.RawMessage

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodGet,
			route: "/user_banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin4\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user4\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "no tag_id",
			httpMethod: http.MethodGet,
			route: "/user_banner?feature_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "no feature_id",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "empty query",
			httpMethod: http.MethodGet,
			route: "/user_banner?",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "non-numeric tag_id",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=a&feature_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "non-numeric feature_id",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=1&feature_id=a",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "banner does not exist",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=1&feature_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "create banner",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [1, 2], \"feature_id\": 1, \"content\": {}, \"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get banner ok",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=1&feature_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &banner,
		},
		{
			caseName: "create inactive banner",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [3, 4], \"feature_id\": 1, \"content\": {}, \"is_active\": false}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get inactive banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=3&feature_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get inactive banner admin",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=3&feature_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "update banner",
			httpMethod: http.MethodPatch,
			route: "/banner/1",
			body: "{\"content\": {\"abc\":2}}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get cached banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=1&feature_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &cachedBanner,
		},
		{
			caseName: "get updated banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=1&feature_id=1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &updatedBanner,
		},
	}

	var token string
	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "get inactive banner" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		if testCase.caseName == "get updated banner" {
			testCase.route += "&use_last_revision=true"
		}

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)

		if testCase.caseName == "get updated banner" {
			t.Log("cached:", cachedBanner)
			t.Log("updated:", updatedBanner)
			require.NotEqual(t, len(cachedBanner), len(updatedBanner))
		}
	}
}

func TestListBanners(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth
	var banners, secondBanners, thirdBanners, afterDeleteBanners []bannerListElement

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodGet,
			route: "/banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin5\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user5\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "user token used",
			httpMethod: http.MethodGet,
			route: "/banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusForbidden,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list ok",
			httpMethod: http.MethodGet,
			route: "/banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &banners,
		},
		{
			caseName: "add banner",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [13, 14], \"feature_id\": 11, \"content\": {\"abc\": [1, 2, 3]}, \"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list ok added one",
			httpMethod: http.MethodGet,
			route: "/banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &secondBanners,
		},
		{
			caseName: "add inactive banner",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [15, 16], \"feature_id\": 11, \"content\": {\"abc\": [1, 2, 3, 4]}, \"is_active\": false}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list ok add inactive",
			httpMethod: http.MethodGet,
			route: "/banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &thirdBanners,
		},
		{
			caseName: "delete banner",
			httpMethod: http.MethodDelete,
			route: "/banner/1",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list ok after delete",
			httpMethod: http.MethodGet,
			route: "/banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &afterDeleteBanners,
		},
	}

	var token string
	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "user token used" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)

		if testCase.caseName == "list ok added one" {
			require.Equal(t, len(banners)+1, len(secondBanners))
		} else if testCase.caseName == "list ok add inactive" {
			require.Equal(t, len(secondBanners)+1, len(thirdBanners))
		} else if testCase.caseName == "list ok after delete" {
			require.Equal(t, len(thirdBanners)-1, len(afterDeleteBanners))
		}
	}
}

func TestListVersions(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth
	var id bannerID
	var versions, secondVersions []versionListElement

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodGet,
			route: "/banner_versions/2",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin6\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user6\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "user token used",
			httpMethod: http.MethodGet,
			route: "/banner_versions/2",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusForbidden,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list versions not found",
			httpMethod: http.MethodGet,
			route: "/banner_versions/150",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "add banner",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [111], \"feature_id\": 111, \"content\": {\"abc\": 1111}, \"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &id,
		},
		{
			caseName: "list versions ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &versions,
		},
		{
			caseName: "update banner",
			httpMethod: http.MethodPatch,
			route: "/banner/",
			body: "{\"content\": {\"abc\": 111}}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list ok added one",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &secondVersions,
		},
		{
			caseName: "delete banner",
			httpMethod: http.MethodDelete,
			route: "/banner/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list not found after deletion",
			httpMethod: http.MethodGet,
			route: "/banner_versions",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
	}

	var token, route string
	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "user token used" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		name := testCase.caseName
		route = testCase.route

		if name == "list versions ok" || name == "update banner" || name == "list ok added one" || name == "delete banner" || name == "list not found after deletion" {
			route += strconv.Itoa(id.ID)
		}

		req, err := buildRequest(testCase.httpMethod, route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)

		if testCase.caseName == "list ok added one" {
			require.Equal(t, len(versions)+1, len(secondVersions))
		}
	}
}

func TestChooseVersion(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth
	var id bannerID
	var versions []versionListElement
	var updatedBanner, chosenBanner json.RawMessage

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodPatch,
			route: "/banner_versions/choose/150",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin7\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user7\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "user token used",
			httpMethod: http.MethodPatch,
			route: "/banner_versions/choose/150",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusForbidden,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "banner not found",
			httpMethod: http.MethodPatch,
			route: "/banner_versions/choose/150?version_id=2",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "add banner",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [111], \"feature_id\": 111, \"content\": {\"abc\": 1111}, \"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &id,
		},
		{
			caseName: "update banner",
			httpMethod: http.MethodPatch,
			route: "/banner/",
			body: "{\"feature_id\": 11111}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "update banner",
			httpMethod: http.MethodPatch,
			route: "/banner/",
			body: "{\"feature_id\": 111111, \"content\":{}}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get updated banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=111&feature_id=111111",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &updatedBanner,
		},
		{
			caseName: "list versions ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &versions,
		},
		{
			caseName: "choose version which does not exist",
			httpMethod: http.MethodPatch,
			route: "/banner_versions/choose/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "choose version ok",
			httpMethod: http.MethodPatch,
			route: "/banner_versions/choose/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get chosen banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=111&feature_id=11111",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &chosenBanner,
		},
		{
			caseName: "add banner",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [111], \"feature_id\": 111, \"content\": {\"abc\": 1111}, \"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "choose version which violates requirements",
			httpMethod: http.MethodPatch,
			route: "/banner_versions/choose/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusConflict,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "non-numeric banner_id and version_id",
			httpMethod: http.MethodPatch,
			route: "/banner_versions/choose/a?version_id=b",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
	}

	var token, route string
	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "user token used" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		name := testCase.caseName
		route = testCase.route

		if name == "choose version which does not exist" || name == "update banner" || name == "list versions ok" || name == "choose version ok" || name == "choose version which violates requirements" {
			route += strconv.Itoa(id.ID)
		}

		if name == "choose version ok" {
			route += "?version_id=" + strconv.Itoa(versions[1].VersionID)
		} else if name == "choose version which violates requirements" {
			route += "?version_id=" + strconv.Itoa(versions[len(versions)-1].VersionID)
		} else if name == "choose version which does not exist" {
			route += "?version_id=" + strconv.Itoa(versions[0].VersionID+1)
		} else if name == "get updated banner" || name == "get chosen banner" {
			route += "&use_last_revision=true"
		}

		req, err := buildRequest(testCase.httpMethod, route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)

		if testCase.caseName == "list versions ok" {
			require.Equal(t, len(versions), 3)
		} else if testCase.caseName == "get chosen banner" {
			require.NotEqual(t, len(updatedBanner), len(chosenBanner))
		}
	}
}

func TestCreateBanner(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth
	var banner json.RawMessage
	var id bannerID
	var versions []versionListElement

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin8\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user8\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "user token used",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusForbidden,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get non-existent banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=42&feature_id=42",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "create banner ok",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [42],\"feature_id\": 42,\"content\": {},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"},},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &id,
		},
		{
			caseName: "get banner ok",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=42&feature_id=42",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &banner,
		},
		{
			caseName: "list versions ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &versions,
		},
	}

	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "list versions ok" {
			testCase.route += strconv.Itoa(id.ID)
		}

		var token string

		if testCase.caseName == "user token used" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		if testCase.caseName == "get banner ok" || testCase.caseName == "get non-existent banner" {
			testCase.route += "&use_last_revision=true"
		}

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
		if testCase.caseName == "get banner ok" {
			require.Equal(t, len(banner), 2)
		} else if testCase.caseName == "list versions ok" {
			require.Equal(t, len(versions), 1)
		}
	}
}

func TestUpdateBanner(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth
	var banner, secondBanner json.RawMessage
	var id bannerID
	var versions, secondVersions []versionListElement

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodPatch,
			route: "/banner/2",
			body: "",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin9\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user9\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "user token used",
			httpMethod: http.MethodPatch,
			route: "/banner/2",
			body: "",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusForbidden,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "no Content-Type",
			httpMethod: http.MethodPatch,
			route: "/banner/2",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "create banner ok",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [420],\"feature_id\": 420,\"content\": {},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &id,
		},
		{
			caseName: "get banner ok",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=420&feature_id=420",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &banner,
		},
		{
			caseName: "list versions ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &versions,
		},
		{
			caseName: "update banner ok",
			httpMethod: http.MethodPatch,
			route: "/banner/",
			body: "{\"feature_id\":421,\"content\":{\"key\":\"value\"}}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list versions after update ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &secondVersions,
		},
		{
			caseName: "get updated banner ok",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=420&feature_id=421",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &secondBanner,
		},
	}

	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "list versions ok" || testCase.caseName == "update banner ok" || testCase.caseName == "list versions after update ok" {
			testCase.route += strconv.Itoa(id.ID)
		}

		var token string

		if testCase.caseName == "user token used" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
		if testCase.caseName == "list versions ok" {
			require.Equal(t, len(versions), 1)
		} else if testCase.caseName == "list versions after update ok" {
			require.Equal(t, 2, len(secondVersions))
		} else if testCase.caseName == "get updated banner ok" {
			require.NotEqual(t, len(banner), len(secondBanner))
		}
	}
}

func TestDeleteBanner(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth
	var banner, secondBanner json.RawMessage
	var id bannerID
	var versions, secondVersions []versionListElement

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodDelete,
			route: "/banner/2",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin10\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user10\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "user token used",
			httpMethod: http.MethodDelete,
			route: "/banner/2",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusForbidden,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "create banner ok",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [500],\"feature_id\": 500,\"content\": {},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &id,
		},
		{
			caseName: "get banner ok",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=500&feature_id=500",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &banner,
		},
		{
			caseName: "list versions ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &versions,
		},
		{
			caseName: "update banner ok",
			httpMethod: http.MethodPatch,
			route: "/banner/",
			body: "{\"feature_id\":600,\"content\":{\"key\":\"value\"}}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list versions after update ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &secondVersions,
		},
		{
			caseName: "get updated banner ok",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=500&feature_id=600",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &secondBanner,
		},
		{
			caseName: "delete non-existent banner",
			httpMethod: http.MethodDelete,
			route: "/banner/1001",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "delete banner ok",
			httpMethod: http.MethodDelete,
			route: "/banner/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get deleted banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=500&feature_id=600",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list versions of deleted banner",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
	}

	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "list versions of deleted banner" || testCase.caseName == "delete banner ok" || testCase.caseName == "list versions ok" || testCase.caseName == "update banner ok" || testCase.caseName == "list versions after update ok" {
			testCase.route += strconv.Itoa(id.ID)
		}

		var token string

		if testCase.caseName == "user token used" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		if testCase.caseName == "get deleted banner" {
			testCase.route += "&use_last_revision=true"
		}

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
		if testCase.caseName == "list versions ok" {
			require.Equal(t, len(versions), 1)
		} else if testCase.caseName == "list versions after update ok" {
			require.Equal(t, 2, len(secondVersions))
		} else if testCase.caseName == "get updated banner ok" {
			require.NotEqual(t, len(banner), len(secondBanner))
		}
	}
}

func TestDeleteBannerByTagOrFeature(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var adminAuthData, userAuthData auth
	var banner, secondBanner json.RawMessage
	var firstID, secondID, thirdID, fourthID, fifthID bannerID
	var versions, secondVersions []versionListElement

	var testTable = []testTableElem {
		{
			caseName: "no auth",
			httpMethod: http.MethodDelete,
			route: "/banner/2",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusUnauthorized,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "register admin",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"admin11\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &adminAuthData,
		},
		{
			caseName: "register user",
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"user11\",\"password\": \"password\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &userAuthData,
		},
		{
			caseName: "user token used",
			httpMethod: http.MethodDelete,
			route: "/banner/2",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusForbidden,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "create banner ok",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [500],\"feature_id\": 500,\"content\": {},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &firstID,
		},
		{
			caseName: "get banner ok",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=500&feature_id=500",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &banner,
		},
		{
			caseName: "list versions ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &versions,
		},
		{
			caseName: "update banner ok",
			httpMethod: http.MethodPatch,
			route: "/banner/",
			body: "{\"feature_id\":600,\"content\":{\"key\":\"value\"}}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list versions after update ok",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &secondVersions,
		},
		{
			caseName: "get updated banner ok",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=500&feature_id=600",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody: &secondBanner,
		},
		{
			caseName: "create banner ok",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [700],\"feature_id\": 700,\"content\": {},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &secondID,
		},
		{
			caseName: "create banner ok",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [800],\"feature_id\": 700,\"content\": {},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &thirdID,
		},
		{
			caseName: "create banner ok",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [850],\"feature_id\": 750,\"content\": {},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &fourthID,
		},
		{
			caseName: "create banner ok",
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [851],\"feature_id\": 750,\"content\": {},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &fifthID,
		},
		{
			caseName: "delete non-existent banners",
			httpMethod: http.MethodDelete,
			route: "/banner?tag_id=999",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "invalid amount of parameters",
			httpMethod: http.MethodDelete,
			route: "/banner?",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "delete banners by feature",
			httpMethod: http.MethodDelete,
			route: "/banner?feature_id=700",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusAccepted,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "delete banner by tag",
			httpMethod: http.MethodDelete,
			route: "/banner?tag_id=500",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusAccepted,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "delete banner by tag and feature",
			httpMethod: http.MethodDelete,
			route: "/banner?tag_id=850&feature_id=750",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusAccepted,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "delete banner by tag and feature",
			httpMethod: http.MethodDelete,
			route: "/banner?tag_id=851&feature_id=750",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusAccepted,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get deleted banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=500&feature_id=600",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "get deleted banner",
			httpMethod: http.MethodGet,
			route: "/user_banner?tag_id=851&feature_id=750",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "list versions of deleted banner",
			httpMethod: http.MethodGet,
			route: "/banner_versions/",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody: nil,
		},
	}

	for _, testCase := range testTable {
		t.Log(testCase.caseName)

		if testCase.caseName == "get deleted banner" {
			testCase.route += "&use_last_revision=true"
			<-time.After(time.Millisecond * 30)
		}

		if testCase.caseName == "list versions of deleted banner" || testCase.caseName == "list versions ok" || testCase.caseName == "update banner ok" || testCase.caseName == "list versions after update ok" {
			testCase.route += strconv.Itoa(firstID.ID)
		}

		var token string

		if testCase.caseName == "user token used" {
			token = userAuthData.Token
		} else {
			token = adminAuthData.Token
		}

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, append(testCase.headers, [2]string{"token", token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
		if testCase.caseName == "list versions ok" {
			require.Equal(t, len(versions), 1)
		} else if testCase.caseName == "list versions after update ok" {
			require.Equal(t, 2, len(secondVersions))
		} else if testCase.caseName == "get updated banner ok" {
			require.NotEqual(t, len(banner), len(secondBanner))
		}
	}
}