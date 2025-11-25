package handlers

import (
	"net/http"

	"github.com/damacus/iron-buckets/internal/services"
	"github.com/damacus/iron-buckets/internal/utils"
	"github.com/labstack/echo/v4"
)

// GetCredentials retrieves and validates credentials from the context
func GetCredentials(c echo.Context) (*services.Credentials, error) {
	val := c.Get(utils.ContextKeyCreds)
	if val == nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	creds, ok := val.(*services.Credentials)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	return creds, nil
}

// GetCredentialsOrRedirect retrieves credentials or redirects to login
func GetCredentialsOrRedirect(c echo.Context) (*services.Credentials, error) {
	creds, err := GetCredentials(c)
	if err != nil {
		return nil, c.Redirect(http.StatusSeeOther, "/login")
	}
	return creds, nil
}

// HTMXRedirect sets the HX-Redirect header and returns a 200 OK response.
// This is used for HTMX requests that should trigger a client-side redirect.
func HTMXRedirect(c echo.Context, url string) error {
	c.Response().Header().Set("HX-Redirect", url)
	return c.NoContent(http.StatusOK)
}
