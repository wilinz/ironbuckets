package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/damacus/iron-buckets/internal/handlers"
	"github.com/damacus/iron-buckets/internal/middleware"
	"github.com/damacus/iron-buckets/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/minio/madmin-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDashboardMetricsJourney(t *testing.T) {
	// 1. Setup
	e := echo.New()
	e.Renderer = &MockRenderer{}
	authService := services.NewAuthService()
	mockFactory := new(MockMinioFactory)
	mockClient := new(MockMinioClient)

	creds := services.Credentials{Endpoint: "play.minio.io:9000", AccessKey: "admin", SecretKey: "password"}
	mockFactory.On("NewAdminClient", creds).Return(mockClient, nil)

	// Mock ServerInfo with detailed metrics
	mockClient.On("ServerInfo", mock.Anything, mock.Anything).Return(madmin.InfoMessage{
		Mode:   "online",
		Region: "us-east-1",
		Servers: []madmin.ServerProperties{
			{
				Endpoint: "minio1:9000",
				Version:  "RELEASE.2024-01-01T00-00-00Z",
				Uptime:   86400, // 1 day in seconds
				Network: map[string]string{
					"minio1:9000": "online",
				},
			},
		},
	}, nil)

	encrypted, _ := authService.EncryptCredentials(creds)
	cookie := &http.Cookie{Name: "IronSeal", Value: encrypted}

	// Routes
	app := e.Group("")
	app.Use(middleware.AuthMiddleware(authService))
	dashboardHandler := handlers.NewDashboardHandler(mockFactory)
	app.GET("/api/server/widget", dashboardHandler.GetServerWidget)

	// Step A: Get Server Widget
	reqWidget := httptest.NewRequest(http.MethodGet, "/api/server/widget", nil)
	reqWidget.AddCookie(cookie)
	recWidget := httptest.NewRecorder()
	e.ServeHTTP(recWidget, reqWidget)

	assert.Equal(t, http.StatusOK, recWidget.Code)
}
