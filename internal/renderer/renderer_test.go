package renderer

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestTemplateRenderer_RenderUnknownTemplate(t *testing.T) {
	r := &TemplateRenderer{
		Templates: make(map[string]*template.Template),
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := r.Render(rec, "nonexistent", nil, c)

	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Contains(t, httpErr.Message, "Template not found")
}

func TestSelfExecutingTemplates_ContainsExpectedTemplates(t *testing.T) {
	expectedTemplates := []string{
		"user_create_modal",
		"bucket_create_modal",
		"folder_create_modal",
		"group_create_modal",
		"drives_widget",
		"storage_widget",
		"users_widget",
		"server_widget",
		"share_link",
		"policy_modal",
		"service_account_create_modal",
		"service_account_created",
		"lifecycle_rules",
		"notifications",
		"object_info",
		"replication",
		"versioning_status",
		"bucket_quota",
		"logs",
	}

	for _, tmpl := range expectedTemplates {
		t.Run(tmpl, func(t *testing.T) {
			if !selfExecutingTemplates[tmpl] {
				t.Errorf("expected %q to be in selfExecutingTemplates", tmpl)
			}
		})
	}
}

func TestSelfExecutingTemplates_DoesNotContainPageTemplates(t *testing.T) {
	pageTemplates := []string{
		"dashboard",
		"drives",
		"users",
		"groups",
		"buckets",
		"browser",
		"settings",
		"login",
	}

	for _, tmpl := range pageTemplates {
		t.Run(tmpl, func(t *testing.T) {
			if selfExecutingTemplates[tmpl] {
				t.Errorf("page template %q should not be in selfExecutingTemplates", tmpl)
			}
		})
	}
}

func TestTemplateRenderer_RenderSelfExecutingTemplate(t *testing.T) {
	// Create a simple template that defines its own block
	tmpl := template.Must(template.New("test_widget").Parse(`{{ define "test_widget" }}Hello {{ .Name }}{{ end }}`))

	r := &TemplateRenderer{
		Templates: map[string]*template.Template{
			"test_widget": tmpl,
		},
	}

	// Temporarily add to selfExecutingTemplates for test
	originalValue := selfExecutingTemplates["test_widget"]
	selfExecutingTemplates["test_widget"] = true
	defer func() {
		if originalValue {
			selfExecutingTemplates["test_widget"] = true
		} else {
			delete(selfExecutingTemplates, "test_widget")
		}
	}()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var buf bytes.Buffer
	err := r.Render(&buf, "test_widget", map[string]interface{}{"Name": "World"}, c)

	assert.NoError(t, err)
	assert.Equal(t, "Hello World", buf.String())
}
