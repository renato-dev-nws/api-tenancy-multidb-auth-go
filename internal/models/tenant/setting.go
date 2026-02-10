package tenant

import (
	"encoding/json"
	"time"
)

// Setting represents a tenant configuration setting
type Setting struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// UpdateSettingRequest represents the request to update a setting
type UpdateSettingRequest struct {
	Value json.RawMessage `json:"value" binding:"required"`
}

// SettingListResponse represents the list of settings
type SettingListResponse struct {
	Settings []Setting `json:"settings"`
}
