package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grqphical/f-stop/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestPhotoUpload(t *testing.T) {
	testPhoto, err := testDataFS.ReadFile("test_data/perlin_noise.png")
	assert.NoError(t, err)

	hasher := sha256.New()
	hasher.Write(testPhoto)

	originalFileHash := hasher.Sum(nil)
	hasher.Reset()

	router := server.NewMockServer()

	// create necessary authentication cookie
	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	assert.NotNil(t, authorizationCookie, "authorizationCookie is nil")

	// upload the photo
	w := httptest.NewRecorder()
	req := newMultipartFileRequest(t, "/api/v1/photo", "test_data/perlin_noise.png")
	req.AddCookie(authorizationCookie)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var respJSON map[string]string
	err = json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	photoID, exists := respJSON["photoId"]
	assert.True(t, exists, "field 'photoId' does not exist on response")

	// make sure the photo's metadata is correct
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/photo/%s", photoID), nil)
	req.AddCookie(authorizationCookie)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respJSON2 = make(map[string]any)
	err = json.NewDecoder(w.Body).Decode(&respJSON2)
	assert.NoError(t, err, "error occurred while decoding JSON")

	assert.Equal(t, len(testPhoto), int(respJSON2["size"].(float64)))
	assert.Equal(t, photoID, respJSON2["id"])
	assert.Equal(t, 0, int(respJSON2["ownerID"].(float64)))
	assert.Equal(t, "image/png", respJSON2["mimeType"])
	assert.Equal(t, fmt.Sprintf("/storage/%s.png", photoID), respJSON2["permalink"])

	// check if the photo itself is correct
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, respJSON2["permalink"].(string), nil)
	req.AddCookie(authorizationCookie)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))

	var respBody bytes.Buffer
	_, err = io.Copy(&respBody, w.Body)
	assert.NoError(t, err)

	io.Copy(hasher, &respBody)

	assert.Equal(t, originalFileHash, hasher.Sum(nil), "file hashes do not match")

}
