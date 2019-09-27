package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)


// This test checks if version-get works and returns expected version string.
func TestVersionGet(t *testing.T) {

	const exp_version = "0.0.1"
	
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/version-get", nil)
	router.ServeHTTP(w, req)

	// Make sure the http status code is 200
	assert.Equal(t, http.StatusOK, w.Code)

	// Make sure the body contains version
	assert.Equal(t, "{\"version\":\"" + exp_version + "\"}\n", w.Body.String())
}
