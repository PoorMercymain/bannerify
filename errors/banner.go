package errors

import "errors"

var (
	ErrTagOrFeatureNotProvided = errors.New("tag_id or/and feature_id was/were not provided")
	ErrTagIsNotANumber = errors.New("provided tag_id is not a number")
	ErrFeatureIsNotANumber = errors.New("provided feature_id is not a number")
	ErrBannerNotFound = errors.New("requested banner not found")
	ErrOffsetIsNotANumber = errors.New("provided offset is not a number")
	ErrLimitIsNotANumber = errors.New("provided limit is not a number")
	ErrLimitNotInRange = errors.New("limit should be in range [1:100]")
	ErrOffsetNotInRange = errors.New("offset should be more than zero")
	ErrFeatureNotInRange = errors.New("feature_id should be more than zero")
	ErrTagNotInRange = errors.New("tag_id should be more than zero")
)