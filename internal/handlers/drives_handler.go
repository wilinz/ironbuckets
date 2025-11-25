package handlers

import (
	"net/http"

	"github.com/damacus/iron-buckets/internal/services"
	"github.com/damacus/iron-buckets/internal/utils"
	"github.com/labstack/echo/v4"
)

type DrivesHandler struct {
	minioFactory services.MinioClientFactory
}

func NewDrivesHandler(minioFactory services.MinioClientFactory) *DrivesHandler {
	return &DrivesHandler{minioFactory: minioFactory}
}

// ListDrives renders the drives page with disk information
func (h *DrivesHandler) ListDrives(c echo.Context) error {
	creds, err := GetCredentialsOrRedirect(c)
	if err != nil {
		return err
	}

	// Connect to MinIO Admin
	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return c.Render(http.StatusOK, "drives", map[string]interface{}{
			"ActiveNav": "drives",
			"Error":     "Failed to connect to MinIO",
		})
	}

	// Fetch Server Info to get disk information
	serverInfo, err := mdm.ServerInfo(c.Request().Context())
	if err != nil {
		return c.Render(http.StatusOK, "drives", map[string]interface{}{
			"ActiveNav": "drives",
			"Error":     "Unable to fetch drive information (admin permissions required)",
		})
	}

	// Extract all disks from all servers
	var allDisks []map[string]interface{}
	onlineCount := 0
	totalCount := 0

	for _, server := range serverInfo.Servers {
		for _, disk := range server.Disks {
			totalCount++
			status := "offline"
			statusColor := "red"

			if disk.State == "ok" {
				onlineCount++
				status = "online"
				statusColor = "emerald"
			}

			usedPercent := 0.0
			if disk.TotalSpace > 0 {
				usedPercent = float64(disk.UsedSpace) / float64(disk.TotalSpace) * 100
			}

			allDisks = append(allDisks, map[string]interface{}{
				"Path":        disk.DrivePath,
				"Endpoint":    server.Endpoint,
				"State":       disk.State,
				"Status":      status,
				"StatusColor": statusColor,
				"TotalSpace":  utils.FormatBytes(disk.TotalSpace),
				"UsedSpace":   utils.FormatBytes(disk.UsedSpace),
				"AvailSpace":  utils.FormatBytes(disk.AvailableSpace),
				"UsedPercent": usedPercent,
				"Healing":     disk.Healing,
				"UUID":        disk.UUID,
			})
		}
	}

	return c.Render(http.StatusOK, "drives", map[string]interface{}{
		"ActiveNav":   "drives",
		"Disks":       allDisks,
		"OnlineCount": onlineCount,
		"TotalCount":  totalCount,
	})
}

// GetDrivesWidget returns drive stats for the dashboard widget
func (h *DrivesHandler) GetDrivesWidget(c echo.Context) error {
	creds, err := GetCredentials(c)
	if err != nil {
		// Return empty widget if not authenticated
		return c.Render(http.StatusOK, "drives_widget", map[string]interface{}{
			"OnlineCount": 0,
			"TotalCount":  0,
			"Status":      "unknown",
		})
	}

	mdm, err := h.minioFactory.NewAdminClient(*creds)
	if err != nil {
		return c.Render(http.StatusOK, "drives_widget", map[string]interface{}{
			"OnlineCount": 0,
			"TotalCount":  0,
			"Status":      "error",
		})
	}

	serverInfo, err := mdm.ServerInfo(c.Request().Context())
	if err != nil {
		return c.Render(http.StatusOK, "drives_widget", map[string]interface{}{
			"OnlineCount": 0,
			"TotalCount":  0,
			"Status":      "error",
		})
	}

	onlineCount := 0
	totalCount := 0

	for _, server := range serverInfo.Servers {
		for _, disk := range server.Disks {
			totalCount++
			if disk.State == "ok" {
				onlineCount++
			}
		}
	}

	status := "healthy"
	if onlineCount < totalCount {
		status = "degraded"
	}
	if onlineCount == 0 {
		status = "offline"
	}

	return c.Render(http.StatusOK, "drives_widget", map[string]interface{}{
		"OnlineCount": onlineCount,
		"TotalCount":  totalCount,
		"Status":      status,
		"Disks":       serverInfo.Servers,
	})
}
