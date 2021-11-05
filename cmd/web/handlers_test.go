package main

import (
	"bytes"
	"net/http"
	"net/url"
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

func TestSignupUser(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/user/signup")
	csrfToken := extractCSRFToken(t, body)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantBody     []byte
	}{
		{
			name:         "Valid submission",
			userName:     "Bob",
			userEmail:    "bob@example.com",
			userPassword: "validPa$$word",
			csrfToken:    csrfToken,
			wantCode:     http.StatusSeeOther,
			wantBody:     nil,
		},
		{
			name:         "Empty name",
			userName:     "",
			userEmail:    "bob@example.com",
			userPassword: "validPa$$word",
			csrfToken:    csrfToken,
			wantCode:     http.StatusOK,
			wantBody:     []byte("This field cannot be blank"),
		},
		{
			name:         "Empty email",
			userName:     "Bob",
			userEmail:    "",
			userPassword: "validPa$$word",
			csrfToken:    csrfToken,
			wantCode:     http.StatusOK,
			wantBody:     []byte("This field cannot be blank"),
		},
		{
			name:         "Empty password",
			userName:     "Bob",
			userEmail:    "bob@example.com",
			userPassword: "",
			csrfToken:    csrfToken,
			wantCode:     http.StatusOK,
			wantBody:     []byte("This field cannot be blank"),
		},
		{
			name:         "Invalid email (incomplete domain)",
			userName:     "Bob",
			userEmail:    "bob@example.",
			userPassword: "validPa$$word",
			csrfToken:    csrfToken,
			wantCode:     http.StatusOK,
			wantBody:     []byte("This field is invalid"),
		},
		{
			name:         "Invalid email (missing @)",
			userName:     "Bob",
			userEmail:    "bobexample.com",
			userPassword: "validPa$$word",
			csrfToken:    csrfToken,
			wantCode:     http.StatusOK,
			wantBody:     []byte("This field is invalid"),
		},
		{
			name:         "Invalid email (missing local part",
			userName:     "Bob",
			userEmail:    "@example.com",
			userPassword: "validPa$$word",
			csrfToken:    csrfToken,
			wantCode:     http.StatusOK,
			wantBody:     []byte("This field is invalid"),
		},
		{
			name:         "Short password",
			userName:     "Bob",
			userEmail:    "bob@example.com",
			userPassword: "Pa$$word",
			csrfToken:    csrfToken,
			wantCode:     http.StatusOK,
			wantBody:     []byte("This field is too short (minimum is 10 characters)"),
		},
		{
			name:         "Duplicate email",
			userName:     "Bob",
			userEmail:    "dupe@example.com",
			userPassword: "validPa$$word",
			csrfToken:    csrfToken,
			wantCode:     http.StatusOK,
			wantBody:     []byte("Address is already in use"),
		},
		{
			name:         "Invalid CSRF Token",
			userName:     "",
			userEmail:    "",
			userPassword: "",
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
			wantBody:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)

			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if !bytes.Contains(body, tt.wantBody) {
				t.Errorf("want body %s to contain %q", body, tt.wantBody)
			}
		})
	}
}

func TestCreateSnippetForm(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name         string
		wantCode     int
		wantLocation string
	}{
		{
			name:         "Unauthenticated",
			wantCode:     http.StatusSeeOther,
			wantLocation: "/user/login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, headers, _ := ts.get(t, "/snippet/create")
			if code != tt.wantCode {
				t.Errorf("want %d; got %d", tt.wantCode, code)
			}
			if headers.Get("Location") != "/user/login" {
				t.Errorf("want %s; got %s", tt.wantLocation, headers.Get("Location"))
			}
		})
	}
}
