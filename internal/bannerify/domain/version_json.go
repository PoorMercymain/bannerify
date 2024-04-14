package domain

import "encoding/json"

type VersionListElement struct {
	VersionID int             `json:"version_id"`
	TagIDs    []int           `json:"tag_ids"`
	FeatureID int             `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
	IsChosen  bool            `json:"is_chosen"`
}
