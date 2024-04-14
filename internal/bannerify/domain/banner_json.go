package domain

import "encoding/json"

type BannerListElement struct {
	BannerID  int             `json:"banner_id"`
	TagIDs    []int           `json:"tag_ids"`
	FeatureID int             `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  bool            `json:"is_active"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

type Banner struct {
	TagIDs    []int           `json:"tag_ids"`
	FeatureID *int            `json:"feature_id"`
	Content   json.RawMessage `json:"content"`
	IsActive  *bool           `json:"is_active"`
}

type BannerID struct {
	ID int `json:"banner_id"`
}
