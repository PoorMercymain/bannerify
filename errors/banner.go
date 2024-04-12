package errors

import "errors"

var (
	ErrTagOrFeatureNotProvided  = errors.New("tag_id or/and feature_id was/were not provided")
	ErrTagIsNotANumber          = errors.New("provided tag_id is not a number")
	ErrFeatureIsNotANumber      = errors.New("provided feature_id is not a number")
	ErrBannerNotFound           = errors.New("requested banner not found")
	ErrOffsetIsNotANumber       = errors.New("provided offset is not a number")
	ErrLimitIsNotANumber        = errors.New("provided limit is not a number")
	ErrLimitNotInRange          = errors.New("limit should be in range [1:100]")
	ErrOffsetNotInRange         = errors.New("offset should not be less than zero")
	ErrFeatureNotInRange        = errors.New("feature_id should be more than zero")
	ErrTagNotInRange            = errors.New("tag_id should be more than zero")
	ErrNoBannerIDProvided       = errors.New("banner_id not found in path")
	ErrBannerIDIsNotANumber     = errors.New("provided banner id is not a number")
	ErrVersionNotFound          = errors.New("version from request does not exist")
	ErrNoVersionIDProvided      = errors.New("version_id not found in query")
	ErrVersionIDIsNotANumber    = errors.New("provided version id is not a number")
	ErrBannerIDNotInRange       = errors.New("banner id should be more than zero")
	ErrVersionIDNotInRange      = errors.New("version_id should be more than zero")
	ErrBannerFieldNotProvided   = errors.New("one or more banner json fields are not provided (tag_ids, feature_id, content and is_active required)")
	ErrBannerTagUniqueViolation = errors.New("feature and tag pair of chosen banners cannot point to different banners")
	ErrNoBannerFieldsProvided   = errors.New("no banner json fields provided (tag_ids or feature_id or content or is_active can be provided)")
	ErrUseLastRevisionNotBool = errors.New("use_last_revision header is not bool (true/false)")
)
