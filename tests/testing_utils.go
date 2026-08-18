package tests

import (
	"bytes"
	"embed"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

//go:embed test_data/*
var testDataFS embed.FS

var authorizationCookie *http.Cookie = nil

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// creates a new account on the given server instance for use in integration tests
func createTestAccount(t *testing.T, router *gin.Engine, username string, email string, password string) {
	formData := map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}
	w := httptest.NewRecorder()
	req := newMultipartRequest(t, "/api/v1/create-account", formData)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "status code not 201 CREATED")

}

// newMultipartRequest builds a POST request populated with form fields and proper headers.
func newMultipartRequest(t *testing.T, target string, fields map[string]string) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			t.Fatalf("Failed to write field %q: %v", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, target, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// newMultipartFileRequest creates an HTTP request with a file upload in a multipart form
func newMultipartFileRequest(t *testing.T, target string, fileName string) *http.Request {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileHandle, err := testDataFS.Open(fileName)
	assert.NoError(t, err)

	fileWriter, err := writer.CreateFormFile("file", fileName)
	assert.NoError(t, err)

	_, err = io.Copy(fileWriter, fileHandle)
	assert.NoError(t, err)

	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, target, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func generateAuthorizationCookie(t *testing.T, router *gin.Engine, email string, password string) {
	if authorizationCookie != nil {
		return
	}

	formData := map[string]string{
		"email":    email,
		"password": password,
	}

	w := httptest.NewRecorder()
	req := newMultipartRequest(t, "/api/v1/login", formData)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Result().Cookies())

	cookie := w.Result().Cookies()[0]
	assert.Equal(t, cookie.Name, "Authorization")
	assert.Equal(t, cookie.HttpOnly, true)

	authorizationCookie = cookie
}
