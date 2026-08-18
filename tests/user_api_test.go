package tests

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
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

func TestUserCreate(t *testing.T) {
	router := server.NewMockServer()

	formData := map[string]string{
		"username": "johndoe",
		"email":    "johndoe@gmail.com",
		"password": "IAmAPassword!",
	}

	// 1. Initial creation succeeds
	w := httptest.NewRecorder()
	req := newMultipartRequest(t, "/api/v1/create-account", formData)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "status code not 201 CREATED")

	// 2. Duplicate account creation fails
	w = httptest.NewRecorder()
	req = newMultipartRequest(t, "/api/v1/create-account", formData)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "status code not 400 BAD REQUEST")

	var respJSON map[string]string
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "UserAlreadyExists", respJSON["errorType"], "incorrect error message received")
}
