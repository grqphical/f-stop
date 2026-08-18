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

var authorizationCookie *http.Cookie

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

func TestUserLogin(t *testing.T) {
	router := server.NewMockServer()

	// create an account to test with
	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")

	// test the login endpoint
	formData := map[string]string{
		"email":    "johndoe@gmail.com",
		"password": "IAmAPassword!",
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

	// test the login endpoint with an incorrect password
	formData = map[string]string{
		"email":    "johndoe@gmail.com",
		"password": "IAmTheWrongPassword!",
	}

	w = httptest.NewRecorder()
	req = newMultipartRequest(t, "/api/v1/login", formData)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	assert.Equal(t, respJSON["errorType"], "InvalidCredentials")

	// test the login endpoint with a non-existent email
	formData = map[string]string{
		"email":    "janedoe@gmail.com",
		"password": "IAmAPassword!",
	}

	w = httptest.NewRecorder()
	req = newMultipartRequest(t, "/api/v1/login", formData)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	err = json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	assert.Equal(t, respJSON["errorType"], "UserNotFound")
}

func TestGetUserInfo(t *testing.T) {
	router := server.NewMockServer()
	w := httptest.NewRecorder()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")

	req, _ := http.NewRequest("GET", "/api/v1/user", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "did not recieve 200 OK")

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	assert.Equal(t, respJSON["id"], 0.0, "ids do not match")
	assert.Equal(t, respJSON["username"], "johndoe", "usernames do not match")
	assert.Equal(t, respJSON["email"], "johndoe@gmail.com", "emails do not match")

}
