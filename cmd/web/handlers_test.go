package main

import (
	"bytes"
	"net/http"
	"testing"
)

func Test_ping(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	if code != http.StatusOK {
		t.Errorf("want %d; got %d", http.StatusOK, code)
	}

	if string(body) != "OK" {
		t.Errorf("want body to equal %q", "OK")
	}
}

func TestShowSnippet(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody []byte
	}{
		{
			name:     "Valid ID",
			urlPath:  "/snippet/1",
			wantCode: http.StatusOK,
			wantBody: []byte("An old silent pond..."),
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/snippet/2",
			wantCode: http.StatusNotFound,
			wantBody: nil,
		},
		{
			name:     "Negative ID",
			urlPath:  "/snippet/-1",
			wantCode: http.StatusNotFound,
			wantBody: nil,
		},
		{
			name:     "Decimal ID",
			urlPath:  "/snippet/1.23",
			wantCode: http.StatusNotFound,
			wantBody: nil,
		},
		{
			name:     "String ID",
			urlPath:  "/snippet/foo",
			wantCode: http.StatusNotFound,
			wantBody: nil,
		},
		{
			name:     "Empty ID",
			urlPath:  "/snippet/",
			wantCode: http.StatusNotFound,
			wantBody: nil,
		},
		{
			name:     "Trailing slash",
			urlPath:  "/snippet/1/",
			wantCode: http.StatusNotFound,
			wantBody: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body to contain %q", tt.wantBody)
			}
		})
	}
}
