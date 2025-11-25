package handlers

import (
	"net/http"

	"github.com/damacus/iron-buckets/internal/services"
	"github.com/labstack/echo/v4"
)

type SettingsHandler struct {
	minioFactory  services.MinioClientFactory
	minioEndpoint string
}

func NewSettingsHandler(minioFactory services.MinioClientFactory, minioEndpoint string) *SettingsHandler {
	return &SettingsHandler{
		minioFactory:  minioFactory,
		minioEndpoint: minioEndpoint,
	}
}

// ShowSettings renders the settings page with server information
func (h *SettingsHandler) ShowSettings(c echo.Context) error {
	creds, err := GetCredentialsOrRedirect(c)
	if err != nil {
		return err
	}

	// Connect to MinIO Admin
	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		// If we can't connect, show settings page with error
		return c.Render(http.StatusOK, "settings", map[string]interface{}{
			"ActiveNav": "settings",
			"Error":     "Failed to connect to MinIO",
			"Endpoint":  h.minioEndpoint,
		})
	}

	data := map[string]interface{}{
		"ActiveNav": "settings",
		"Endpoint":  h.minioEndpoint,
	}

	// Fetch Server Info
	serverInfo, err := mdm.ServerInfo(c.Request().Context())
	if err != nil {
		// User might not have admin permissions
		data["Error"] = "Unable to fetch server information (admin permissions required)"
		return c.Render(http.StatusOK, "settings", data)
	}
	data["ServerInfo"] = serverInfo

	// Fetch Storage Info (DataUsageInfo)
	storageInfo, err := mdm.DataUsageInfo(c.Request().Context())
	if err == nil {
		data["StorageInfo"] = storageInfo
	}

	// Fetch Server Config (if available)
	config, err := mdm.GetConfig(c.Request().Context())
	if err == nil {
		data["Config"] = string(config)
	}

	return c.Render(http.StatusOK, "settings", data)
}

// RestartService handles MinIO service restart
func (h *SettingsHandler) RestartService(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	// Restart the service
	if err := mdm.ServiceRestart(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to restart service: "+err.Error())
	}

	return HTMXRedirect(c, "/settings")
}

// GetLogs returns recent server logs
func (h *SettingsHandler) GetLogs(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		return err
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to MinIO")
	}

	// Get last 100 log lines
	logCh := mdm.GetLogs(c.Request().Context(), "", 100, "all")

	var logs []map[string]interface{}
	for logInfo := range logCh {
		if logInfo.Err != nil {
			continue
		}
		logs = append(logs, map[string]interface{}{
			"Node":    logInfo.NodeName,
			"Message": logInfo.ConsoleMsg,
		})
	}

	return c.Render(http.StatusOK, "logs", map[string]interface{}{
		"Logs":    logs,
		"HasLogs": len(logs) > 0,
	})
}
