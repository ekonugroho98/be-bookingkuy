package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ekonugroho98/be-bookingkuy/internal/shared/jwt"
	"github.com/stretchr/testify/require"
)

// TestServer wraps httptest.Server with additional helper methods
type TestServer struct {
	Server *httptest.Server
	T      *testing.T
}

// NewTestServer creates a new test server with the given handler
func NewTestServer(t *testing.T, handler http.Handler) *TestServer {
	t.Helper()
	return &TestServer{
		Server: httptest.NewServer(handler),
		T:      t,
	}
}

// Close closes the test server
func (ts *TestServer) Close() {
	ts.Server.Close()
}

// URL returns the base URL of the test server
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// Do performs an HTTP request with the given method, path, and body
func (ts *TestServer) Do(method, path string, body interface{}, headers map[string]string) *http.Response {
	ts.T.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(ts.T, err)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, ts.Server.URL+path, reqBody)
	require.NoError(ts.T, err)

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Perform request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(ts.T, err)

	return resp
}

// Get performs a GET request
func (ts *TestServer) Get(path string, headers map[string]string) *http.Response {
	ts.T.Helper()
	return ts.Do(http.MethodGet, path, nil, headers)
}

// Post performs a POST request
func (ts *TestServer) Post(path string, body interface{}, headers map[string]string) *http.Response {
	ts.T.Helper()
	return ts.Do(http.MethodPost, path, body, headers)
}

// Put performs a PUT request
func (ts *TestServer) Put(path string, body interface{}, headers map[string]string) *http.Response {
	ts.T.Helper()
	return ts.Do(http.MethodPut, path, body, headers)
}

// Delete performs a DELETE request
func (ts *TestServer) Delete(path string, headers map[string]string) *http.Response {
	ts.T.Helper()
	return ts.Do(http.MethodDelete, path, nil, headers)
}

// DecodeJSON decodes the response body into the given interface
func (ts *TestServer) DecodeJSON(resp *http.Response, v interface{}) {
	ts.T.Helper()
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(v)
	require.NoError(ts.T, err)
}

// GetStatusCode returns the status code of the response
func (ts *TestServer) GetStatusCode(resp *http.Response) int {
	ts.T.Helper()
	return resp.StatusCode
}

// GetResponseBody returns the response body as string
func (ts *TestServer) GetResponseBody(resp *http.Response) string {
	ts.T.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(ts.T, err)
	return string(body)
}

// AuthHelper provides helper methods for authentication in tests
type AuthHelper struct {
	testServer *TestServer
	jwtManager *jwt.Manager
}

// NewAuthHelper creates a new auth helper
func NewAuthHelper(testServer *TestServer, jwtSecret string) *AuthHelper {
	return &AuthHelper{
		testServer: testServer,
		jwtManager: jwt.NewManager(jwtSecret),
	}
}

// RegisterAndLogin registers a new user and returns the auth token
func (ah *AuthHelper) RegisterAndLogin(email, password, name string) string {
	ah.testServer.T.Helper()

	// Register
	registerReq := map[string]string{
		"email":    email,
		"password": password,
		"name":     name,
	}
	resp := ah.testServer.Post("/auth/register", registerReq, nil)
	require.Equal(ah.testServer.T, http.StatusCreated, ah.testServer.GetStatusCode(resp))

	// Login
	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}
	resp = ah.testServer.Post("/auth/login", loginReq, nil)
	require.Equal(ah.testServer.T, http.StatusOK, ah.testServer.GetStatusCode(resp))

	var loginResp struct {
		Token string `json:"token"`
		User  struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"user"`
	}
	ah.testServer.DecodeJSON(resp, &loginResp)

	return loginResp.Token
}

// GenerateTestToken generates a valid JWT token for testing
func (ah *AuthHelper) GenerateTestToken(userID, email string) string {
	ah.testServer.T.Helper()

	token, err := ah.jwtManager.GenerateToken(userID, email)
	require.NoError(ah.testServer.T, err)
	return token
}

// GetAuthHeaders returns headers with authorization token
func (ah *AuthHelper) GetAuthHeaders(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// testContext is a reusable context for tests
var testContext = context.Background()

