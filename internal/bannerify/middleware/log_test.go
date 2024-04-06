package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testLogRouter(t *testing.T) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("GET /logs", Log(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	mux.Handle("GET /logs-with-write", Log(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("abc"))
		require.NoError(t, err)
	})))

	return mux
}

func TestLog(t *testing.T) {
	ts := httptest.NewServer(testLogRouter(t))

	defer ts.Close()

	var testTable = []struct {
		endpoint      string
		method        string
		content       string
		code          int
		body          string
		authorization string
	}{
		{
			"/logs",
			http.MethodGet,
			"",
			http.StatusOK,
			"",
			"",
		},
		{
			"/logs-with-write",
			http.MethodGet,
			"",
			http.StatusOK,
			"",
			"",
		},
	}

	for _, testCase := range testTable {
		resp := request(t, ts, testCase.code, testCase.method, testCase.content, testCase.body, testCase.endpoint, testCase.authorization)
		resp.Body.Close()
	}
}
