package handlers

import (
	"net/http"

	"github.com/edsonmichaque/bazaruto/internal/version"
)

// VersionHandler handles version information requests
type VersionHandler struct{}

// NewVersionHandler creates a new version handler
func NewVersionHandler() *VersionHandler {
	return &VersionHandler{}
}

// GetVersion returns version information
func (h *VersionHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	info := version.Get()
	WriteJSONIgnoreError(w, http.StatusOK, info)
}
