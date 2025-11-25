package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/damacus/iron-buckets/internal/services"
	"github.com/damacus/iron-buckets/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetCredentials_WithValidCredentials(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedCreds := &services.Credentials{
		Endpoint:  "localhost:9000",
		AccessKey: "admin",
		SecretKey: "password",
	}
	c.Set(utils.ContextKeyCreds, expectedCreds)

	creds, err := GetCredentials(c)

	assert.NoError(t, err)
	assert.NotNil(t, creds)
	assert.Equal(t, expectedCreds.Endpoint, creds.Endpoint)
	assert.Equal(t, expectedCreds.AccessKey, creds.AccessKey)
	assert.Equal(t, expectedCreds.SecretKey, creds.SecretKey)
}

func TestGetCredentials_WithoutCredentials(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	creds, err := GetCredentials(c)

	assert.Error(t, err)
	assert.Nil(t, creds)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestGetCredentials_WithWrongType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set wrong type in context
	c.Set(utils.ContextKeyCreds, "not-credentials")

	creds, err := GetCredentials(c)

	assert.Error(t, err)
	assert.Nil(t, creds)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
}

func TestGetCredentialsOrRedirect_WithValidCredentials(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedCreds := &services.Credentials{
		Endpoint:  "localhost:9000",
		AccessKey: "admin",
		SecretKey: "password",
	}
	c.Set(utils.ContextKeyCreds, expectedCreds)

	creds, err := GetCredentialsOrRedirect(c)

	assert.NoError(t, err)
	assert.NotNil(t, creds)
	assert.Equal(t, expectedCreds.Endpoint, creds.Endpoint)
}

func TestGetCredentialsOrRedirect_WithoutCredentials(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	creds, err := GetCredentialsOrRedirect(c)

	assert.Nil(t, creds)
	// err is nil because redirect returns nil error
	assert.Nil(t, err)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/login", rec.Header().Get("Location"))
}

func TestHTMXRedirect(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := HTMXRedirect(c, "/dashboard")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "/dashboard", rec.Header().Get("HX-Redirect"))
}

func TestHTMXRedirect_WithQueryParams(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := HTMXRedirect(c, "/buckets/my-bucket?prefix=folder/")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "/buckets/my-bucket?prefix=folder/", rec.Header().Get("HX-Redirect"))
}
