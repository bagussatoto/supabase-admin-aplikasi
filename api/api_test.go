package api

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth(t *testing.T) {
	api := NewAPIWithVersion(&Config{
		JwtSecret: "awdawdawdawdawdaw",
	}, "0.0")
	ts := httptest.NewServer(api.handler)
	defer ts.Close()

	if response, body := testRequest(t, ts, "GET", "/test/", nil, false); response.StatusCode != 401 || body != "\"\"" {
		t.Fatalf("Unauth request should've failed %+v", body)
	}
	if response, body := testRequest(t, ts, "GET", "/test/", nil, true); body != "\"hello\"" || response.StatusCode != 200 {
		t.Fatalf("authed request should've succeeded %+v %+v", response.StatusCode, body)
	}
}

// we only expect this to work on linux
func TestMetrics(t *testing.T) {
	api := NewAPIWithVersion(&Config{}, "0.0")
	ts := httptest.NewServer(api.handler)
	defer ts.Close()

	if response, _ := testRequest(t, ts, "GET", "/metrics", nil, false); response.StatusCode != 200 {
		t.Fatalf("Metrics request should've succeeded %+v", response.StatusCode)
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader, auth bool) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	if auth {
		req.Header.Set("apikey", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoic3VwYWJhc2VfYWRtaW4ifQ.veeAYq7d22dUiUe7cfQDvnZULmLJwUiB2neF_zTcD94")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}