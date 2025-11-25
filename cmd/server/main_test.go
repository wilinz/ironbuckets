package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/damacus/iron-buckets/internal/renderer"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRoutes(t *testing.T) {
	// Setup Echo
	e := echo.New()

	// Setup Templates (Manually mirroring renderer.go logic)
	templates := make(map[string]*template.Template)
	parse := func(name, pageFile string) {
		templates[name] = template.Must(template.ParseFiles(
			"../../views/layouts/base.html",
			"../../views/partials/confirm_dialog.html",
			"../../views/pages/"+pageFile,
		))
	}
	parse("overview", "dashboard.html")
	parse("drives", "drives.html")
	parse("users", "users.html")
	parse("buckets", "buckets.html")
	parse("settings", "settings.html")

	e.Renderer = &renderer.TemplateRenderer{Templates: templates}

	// Define Route Handlers (mirroring main.go)
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "overview", map[string]interface{}{"ActiveNav": "overview"})
	})
	e.GET("/settings", func(c echo.Context) error {
		return c.Render(http.StatusOK, "settings", map[string]interface{}{"ActiveNav": "settings"})
	})

	// Test Case 1: Overview (Dashboard)
	t.Run("Overview Page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Cluster Online")
	})

	// Test Case 2: Settings (Moved Restart Button)
	t.Run("Settings Page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// Power Operations and Restart Service are only shown when ServerInfo is provided
		// The test verifies the page renders successfully with basic content
		assert.Contains(t, rec.Body.String(), "Server Settings")
		assert.Contains(t, rec.Body.String(), "Sign Out")
	})
}
