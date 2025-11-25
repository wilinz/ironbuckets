// Package utils provides shared utility functions
package utils

import "fmt"

// FormatBytes converts bytes to human-readable format (e.g., "1.5 GB")
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatFileSize converts file size (int64) to human-readable format
func FormatFileSize(size int64) string {
	if size < 0 {
		return "0 B"
	}
	return FormatBytes(uint64(size))
}
