// Package models contains data structures used across handlers
package models

import "time"

// ObjectInfo represents an object with display metadata
type ObjectInfo struct {
	Key           string
	DisplayName   string
	Size          int64
	FormattedSize string
	LastModified  time.Time
	ContentType   string
	IsImage       bool
	IsText        bool
	IsVideo       bool
	IsArchive     bool
	IsPreviewable bool
}

// FolderInfo represents a folder (common prefix)
type FolderInfo struct {
	Name   string
	Prefix string
}

// Breadcrumb for navigation
type Breadcrumb struct {
	Name string
	Path string
}
