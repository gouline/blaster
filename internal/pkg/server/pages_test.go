package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleIndexAuthenticated(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/", nil)
	r.Authenticate("1", "!!ACME!!")

	if assert.NoError(t, r.Server.handleIndex(r.Context)) {
		assert.Equal(t, http.StatusOK, r.Response.Code)
		assert.Contains(t, r.Response.Body.String(), "!!ACME!!")
	}
}

func TestHandleIndexUnauthenticated(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/", nil)

	if assert.NoError(t, r.Server.handleIndex(r.Context)) {
		assert.Equal(t, http.StatusOK, r.Response.Code)
		assert.Contains(t, r.Response.Body.String(), "<body>")
	}
}

func TestHandleNotFound(t *testing.T) {
	r := newRequestTester(http.MethodGet, "/", nil)

	if assert.NoError(t, r.Server.handleNotFound(r.Context)) {
		assert.Equal(t, http.StatusNotFound, r.Response.Code)
		assert.Contains(t, r.Response.Body.String(), "<body>")
	}
}
