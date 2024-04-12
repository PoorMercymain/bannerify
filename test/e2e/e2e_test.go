package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/require"
)

type e2eConfig struct {
	ServicePort int    `env:"SERVICE_PORT" envDefault:"8080"`
	ServiceHost string `env:"SERVICE_HOST" envDefault:"bannerify-test"`
}

type auth struct {
	Token string `json:"token"`
}

type testTableElem struct {
	httpMethod string
	route string
	body string
	headers [][2]string
	expectedStatus int
	requireParsing bool
	parsedBody interface{}
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

func TestCreateBanner(t *testing.T) {
	cfg := e2eConfig{}
	if err := env.Parse(&cfg); err != nil {
		t.Fatalf("Failed to parse env: %v", err)
	}

	client := http.Client{}
	var authData auth

	var testTable = []testTableElem {
		{
			httpMethod: http.MethodPost,
			route: "/register",
			body: "{\"login\": \"aboboss\",\"password\": \"abe\"}",
			headers: [][2]string{{"Content-Type", "application/json"}, {"admin", "true"}},
			expectedStatus: http.StatusCreated,
			requireParsing: true,
			parsedBody: &authData,
		},
		{
			httpMethod: http.MethodPost,
			route: "/banner",
			body: "{\"tag_ids\": [0],\"feature_id\": 0,\"content\": {\"title\": \"some_title\",\"text\": \"some_text\",\"url\": \"some_url\"},\"is_active\": true}",
			headers: [][2]string{{"Content-Type", "application/json"},},
			expectedStatus: http.StatusCreated,
			requireParsing: false,
			parsedBody: nil,
		},
	}

	for _, testCase := range testTable {
		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, append(testCase.headers, [2]string{"token", authData.Token}), cfg)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
	}
}