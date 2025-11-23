package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/111zxc/pr-review-service/internal/domain"
)

func POST(t *testing.T, url string, body any) *http.Response {
	t.Helper()

	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("error marshalling post request: %v", err)
	}
	return POST_RAW(t, url, b)
}

func POST_RAW(t *testing.T, url string, b []byte) *http.Response { //nolint:stylecheck
	t.Helper()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("POST failed: %v", err)
	}
	return resp
}

func GET(t *testing.T, url string) *http.Response {
	t.Helper()

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	return resp
}

func ExpectStatus(t *testing.T, resp *http.Response, status int) {
	t.Helper()
	if resp.StatusCode != status {
		t.Fatalf("expected %d, got %d", status, resp.StatusCode)
	}
}

func ExpectErrorCode(t *testing.T, resp *http.Response, code string) {
	t.Helper()

	var er domain.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if er.Error.Code != code {
		t.Fatalf("expected error code %s, got %s", code, er.Error.Code)
	}
}
