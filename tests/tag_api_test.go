package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/grqphical/f-stop/internal/server"
	"github.com/stretchr/testify/assert"
)

// newMultipartMethodRequest builds a request with the given method populated
// with multipart form fields. Needed because newMultipartRequest only builds
// POST requests, while tag rename uses PUT.
func newMultipartMethodRequest(t *testing.T, method string, target string, fields map[string]string) *http.Request {
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

	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// loginWithCredentials logs in and returns the auth cookie directly, without
// touching the shared authorizationCookie global. Used for second users in
// ownership tests.
func loginWithCredentials(t *testing.T, router *gin.Engine, email string, password string) *http.Cookie {
	t.Helper()

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
	assert.Equal(t, "Authorization", cookie.Name)

	return cookie
}

// createTagAssumeSuccess creates a tag and returns its ID.
func createTagAssumeSuccess(t *testing.T, router *gin.Engine, cookie *http.Cookie, name string) int {
	t.Helper()

	w := httptest.NewRecorder()
	req := newMultipartMethodRequest(t, http.MethodPost, "/api/v1/tags", map[string]string{"name": name})
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	tagID, exists := respJSON["tagId"].(float64)
	assert.True(t, exists, "field 'tagId' does not exist on response")

	return int(tagID)
}

// uploadPhotoAssumeSuccess uploads the test photo and returns its ID.
func uploadPhotoAssumeSuccess(t *testing.T, router *gin.Engine, cookie *http.Cookie) string {
	t.Helper()

	w := httptest.NewRecorder()
	req := newMultipartFileRequest(t, "/api/v1/photo", "test_data/perlin_noise.png")
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	photoID, exists := respJSON["photoId"].(string)
	assert.True(t, exists, "field 'photoId' does not exist on response")

	return photoID
}

// assignTagsAssumeSuccess assigns tags to a photo via PATCH.
func assignTagsAssumeSuccess(t *testing.T, router *gin.Engine, cookie *http.Cookie, photoID string, tagIDs []int) {
	t.Helper()

	payload, err := json.Marshal(map[string]any{"tags": tagIDs})
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/photo/%s/tags", photoID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// getPhotoTags decodes the {"tags": [...]} response for a photo.
func getPhotoTags(t *testing.T, router *gin.Engine, cookie *http.Cookie, photoID string) []any {
	t.Helper()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/photo/%s/tags", photoID), nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	tags, exists := respJSON["tags"].([]any)
	assert.True(t, exists, "field 'tags' does not exist on response")

	return tags
}

// getTagPhotos decodes the {"photos": [...]} response for a tag.
func getTagPhotos(t *testing.T, router *gin.Engine, cookie *http.Cookie, tagID int) []any {
	t.Helper()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tags/%d/photos", tagID), nil)
	req.AddCookie(cookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	photos, exists := respJSON["photos"].([]any)
	assert.True(t, exists, "field 'photos' does not exist on response")

	return photos
}

func TestTagCreate(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	assert.NotNil(t, authorizationCookie, "authorizationCookie is nil")

	// creating a tag succeeds and returns its ID
	tagID := createTagAssumeSuccess(t, router, authorizationCookie, "vacation")
	assert.GreaterOrEqual(t, tagID, 0)

	// unauthenticated creation fails
	w := httptest.NewRecorder()
	req := newMultipartMethodRequest(t, http.MethodPost, "/api/v1/tags", map[string]string{"name": "noauth"})
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTagByID(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	tagID := createTagAssumeSuccess(t, router, authorizationCookie, "birthday")

	// fetching the tag by ID returns it
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tags/%d", tagID), nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	assert.Equal(t, float64(tagID), respJSON["id"])
	assert.Equal(t, "birthday", respJSON["name"])

	// non-numeric ID fails
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tags/notanid", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errJSON map[string]any
	err = json.NewDecoder(w.Body).Decode(&errJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "InvalidID", errJSON["errorType"])

	// unknown ID fails
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tags/9999", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	errJSON = make(map[string]any)
	err = json.NewDecoder(w.Body).Decode(&errJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "TagNotFound", errJSON["errorType"])

	// unauthenticated fetch fails
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tags/%d", tagID), nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTagByName(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	tagID := createTagAssumeSuccess(t, router, authorizationCookie, "sunset")

	// fetching the tag by name returns it
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tags/names/sunset", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	assert.Equal(t, float64(tagID), respJSON["id"])
	assert.Equal(t, "sunset", respJSON["name"])

	// unknown name fails
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tags/names/doesnotexist", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errJSON map[string]any
	err = json.NewDecoder(w.Body).Decode(&errJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "TagNotFound", errJSON["errorType"])
}

func TestListUserTags(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	// no tags initially
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tags", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Empty(t, respJSON["tags"], "tags is not empty")

	// create two tags
	createTagAssumeSuccess(t, router, authorizationCookie, "family")
	createTagAssumeSuccess(t, router, authorizationCookie, "friends")

	// both tags are listed
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tags", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	respJSON = make(map[string]any)
	err = json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")

	tags, exists := respJSON["tags"].([]any)
	assert.True(t, exists, "field 'tags' does not exist on response")
	assert.Len(t, tags, 2)

	names := map[string]bool{}
	for _, tag := range tags {
		tagMap, ok := tag.(map[string]any)
		assert.True(t, ok, "tag is not an object")
		names[tagMap["name"].(string)] = true
	}
	assert.True(t, names["family"], "tag 'family' missing from response")
	assert.True(t, names["friends"], "tag 'friends' missing from response")
}

func TestRenameTag(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	tagID := createTagAssumeSuccess(t, router, authorizationCookie, "oldname")

	// renaming succeeds
	w := httptest.NewRecorder()
	req := newMultipartMethodRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tags/%d", tagID), map[string]string{"name": "newname"})
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// the new name is returned when fetching
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tags/%d", tagID), nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&respJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "newname", respJSON["name"])

	// non-numeric ID fails
	w = httptest.NewRecorder()
	req = newMultipartMethodRequest(t, http.MethodPut, "/api/v1/tags/notanid", map[string]string{"name": "nope"})
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// another user cannot rename the tag
	createTestAccount(t, router, "janedoe", "janedoe@gmail.com", "IAmAPassword!")
	janeCookie := loginWithCredentials(t, router, "janedoe@gmail.com", "IAmAPassword!")

	w = httptest.NewRecorder()
	req = newMultipartMethodRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/tags/%d", tagID), map[string]string{"name": "hijacked"})
	req.AddCookie(janeCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteTag(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	tagID := createTagAssumeSuccess(t, router, authorizationCookie, "todelete")

	// deleting the tag succeeds
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/tags/%d", tagID), nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// fetching the deleted tag fails
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tags/%d", tagID), nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// non-numeric ID fails
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/tags/notanid", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// another user cannot delete the tag
	otherTagID := createTagAssumeSuccess(t, router, authorizationCookie, "keepme")
	createTestAccount(t, router, "janedoe", "janedoe@gmail.com", "IAmAPassword!")
	janeCookie := loginWithCredentials(t, router, "janedoe@gmail.com", "IAmAPassword!")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/tags/%d", otherTagID), nil)
	req.AddCookie(janeCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAssignAndGetPhotoTags(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	photoID := uploadPhotoAssumeSuccess(t, router, authorizationCookie)
	tagID := createTagAssumeSuccess(t, router, authorizationCookie, "landscape")

	// no tags assigned initially
	assert.Empty(t, getPhotoTags(t, router, authorizationCookie, photoID), "photo should have no tags initially")

	// assigning a tag succeeds
	assignTagsAssumeSuccess(t, router, authorizationCookie, photoID, []int{tagID})

	// the assigned tag is returned
	tags := getPhotoTags(t, router, authorizationCookie, photoID)
	assert.Len(t, tags, 1)

	tagMap, ok := tags[0].(map[string]any)
	assert.True(t, ok, "tag is not an object")
	assert.Equal(t, float64(tagID), tagMap["id"])
	assert.Equal(t, "landscape", tagMap["name"])

	// malformed JSON fails
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/photo/%s/tags", photoID), bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&errJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "InvalidRequest", errJSON["errorType"])

	// unauthenticated assignment fails
	payload, err := json.Marshal(map[string]any{"tags": []int{tagID}})
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/photo/%s/tags", photoID), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetTagPhotos(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	tagID := createTagAssumeSuccess(t, router, authorizationCookie, "album")
	photoID1 := uploadPhotoAssumeSuccess(t, router, authorizationCookie)
	photoID2 := uploadPhotoAssumeSuccess(t, router, authorizationCookie)

	// no photos associated initially
	assert.Empty(t, getTagPhotos(t, router, authorizationCookie, tagID), "tag should have no photos initially")

	// assign the tag to both photos
	assignTagsAssumeSuccess(t, router, authorizationCookie, photoID1, []int{tagID})
	assignTagsAssumeSuccess(t, router, authorizationCookie, photoID2, []int{tagID})

	// both photos are returned
	photos := getTagPhotos(t, router, authorizationCookie, tagID)
	assert.Len(t, photos, 2)

	photoIDs := map[string]bool{}
	for _, photo := range photos {
		photoMap, ok := photo.(map[string]any)
		assert.True(t, ok, "photo is not an object")
		id, ok := photoMap["id"].(string)
		assert.True(t, ok, "photo 'id' is not a string")
		photoIDs[id] = true
	}
	assert.True(t, photoIDs[photoID1], "first photo missing from response")
	assert.True(t, photoIDs[photoID2], "second photo missing from response")

	// non-numeric tag ID fails
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/tags/notanid/photos", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errJSON map[string]any
	err := json.NewDecoder(w.Body).Decode(&errJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "InvalidID", errJSON["errorType"])

	// unknown tag ID fails
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/tags/9999/photos", nil)
	req.AddCookie(authorizationCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	errJSON = make(map[string]any)
	err = json.NewDecoder(w.Body).Decode(&errJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "TagNotFound", errJSON["errorType"])

	// another user cannot list photos for the tag
	createTestAccount(t, router, "janedoe", "janedoe@gmail.com", "IAmAPassword!")
	janeCookie := loginWithCredentials(t, router, "janedoe@gmail.com", "IAmAPassword!")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/tags/%d/photos", tagID), nil)
	req.AddCookie(janeCookie)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRemovePhotoTags(t *testing.T) {
	router := server.NewMockServer()

	createTestAccount(t, router, "johndoe", "johndoe@gmail.com", "IAmAPassword!")
	generateAuthorizationCookie(t, router, "johndoe@gmail.com", "IAmAPassword!")

	photoID := uploadPhotoAssumeSuccess(t, router, authorizationCookie)
	tagID1 := createTagAssumeSuccess(t, router, authorizationCookie, "keep")
	tagID2 := createTagAssumeSuccess(t, router, authorizationCookie, "remove")

	assignTagsAssumeSuccess(t, router, authorizationCookie, photoID, []int{tagID1, tagID2})
	assert.Len(t, getPhotoTags(t, router, authorizationCookie, photoID), 2)

	removeTags := func(cookie *http.Cookie, targetPhotoID string, body []byte) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/photo/%s/tags", targetPhotoID), bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			req.AddCookie(cookie)
		}
		router.ServeHTTP(w, req)
		return w
	}

	// removing one tag leaves the other assigned
	payload, err := json.Marshal(map[string]any{"tags": []int{tagID2}})
	assert.NoError(t, err)

	w := removeTags(authorizationCookie, photoID, payload)
	assert.Equal(t, http.StatusOK, w.Code)

	remaining := getPhotoTags(t, router, authorizationCookie, photoID)
	assert.Len(t, remaining, 1)

	remainingMap, ok := remaining[0].(map[string]any)
	assert.True(t, ok, "tag is not an object")
	assert.Equal(t, float64(tagID1), remainingMap["id"])

	// the removed tag no longer lists the photo
	assert.Empty(t, getTagPhotos(t, router, authorizationCookie, tagID2), "removed tag should have no photos")

	// removing the last tag leaves none assigned
	payload, err = json.Marshal(map[string]any{"tags": []int{tagID1}})
	assert.NoError(t, err)

	w = removeTags(authorizationCookie, photoID, payload)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, getPhotoTags(t, router, authorizationCookie, photoID), "photo should have no tags left")

	// removing tags from an unknown photo fails
	payload, err = json.Marshal(map[string]any{"tags": []int{tagID1}})
	assert.NoError(t, err)

	w = removeTags(authorizationCookie, "nonexistent-photo-id", payload)
	assert.Equal(t, http.StatusNotFound, w.Code)

	var errJSON map[string]any
	err = json.NewDecoder(w.Body).Decode(&errJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "PhotoNotFound", errJSON["errorType"])

	// malformed JSON fails
	w = removeTags(authorizationCookie, photoID, []byte("{invalid"))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	errJSON = make(map[string]any)
	err = json.NewDecoder(w.Body).Decode(&errJSON)
	assert.NoError(t, err, "error occurred while decoding JSON")
	assert.Equal(t, "InvalidRequest", errJSON["errorType"])

	// unauthenticated removal fails
	payload, err = json.Marshal(map[string]any{"tags": []int{tagID1}})
	assert.NoError(t, err)

	w = removeTags(nil, photoID, payload)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// another user cannot remove tags from the photo
	assignTagsAssumeSuccess(t, router, authorizationCookie, photoID, []int{tagID1})
	createTestAccount(t, router, "janedoe", "janedoe@gmail.com", "IAmAPassword!")
	janeCookie := loginWithCredentials(t, router, "janedoe@gmail.com", "IAmAPassword!")

	payload, err = json.Marshal(map[string]any{"tags": []int{tagID1}})
	assert.NoError(t, err)

	w = removeTags(janeCookie, photoID, payload)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
