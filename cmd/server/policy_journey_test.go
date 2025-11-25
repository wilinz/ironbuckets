package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/damacus/iron-buckets/internal/handlers"
	"github.com/damacus/iron-buckets/internal/middleware"
	"github.com/damacus/iron-buckets/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPolicyManagementJourney(t *testing.T) {
	// 1. Setup
	e := echo.New()
	e.Renderer = &MockRenderer{}
	authService := services.NewAuthService()
	mockFactory := new(MockMinioFactory)
	mockClient := new(MockMinioClient)

	creds := services.Credentials{Endpoint: "play.minio.io:9000", AccessKey: "admin", SecretKey: "password"}
	mockFactory.On("NewAdminClient", creds).Return(mockClient, nil)

	// Mock listing policies
	mockClient.On("ListCannedPolicies", mock.Anything).Return(map[string]json.RawMessage{
		"readwrite":    json.RawMessage(`{"Version":"2012-10-17"}`),
		"readonly":     json.RawMessage(`{"Version":"2012-10-17"}`),
		"consoleAdmin": json.RawMessage(`{"Version":"2012-10-17"}`),
	}, nil)

	// Mock attaching policy to user
	mockClient.On("SetPolicy", mock.Anything, "readwrite", "testuser", false).Return(nil)

	encrypted, _ := authService.EncryptCredentials(creds)
	cookie := &http.Cookie{Name: "IronSeal", Value: encrypted}

	// Routes
	app := e.Group("")
	app.Use(middleware.AuthMiddleware(authService))
	usersHandler := handlers.NewUsersHandler(mockFactory)
	app.GET("/policies", usersHandler.ListPolicies)
	app.POST("/users/:accessKey/policy", usersHandler.AttachPolicy)

	// Step A: List Policies
	reqList := httptest.NewRequest(http.MethodGet, "/policies", nil)
	reqList.AddCookie(cookie)
	recList := httptest.NewRecorder()
	e.ServeHTTP(recList, reqList)
	assert.Equal(t, http.StatusOK, recList.Code)

	// Step B: Attach Policy to User
	form := make(url.Values)
	form.Set("policy", "readwrite")

	reqAttach := httptest.NewRequest(http.MethodPost, "/users/testuser/policy", strings.NewReader(form.Encode()))
	reqAttach.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	reqAttach.AddCookie(cookie)
	recAttach := httptest.NewRecorder()
	e.ServeHTTP(recAttach, reqAttach)
	assert.Equal(t, http.StatusOK, recAttach.Code)
}
