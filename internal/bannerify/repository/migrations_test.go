package repository

import (
	"errors"
	"testing"

	"github.com/PoorMercymain/bannerify/internal/bannerify/repository/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockMigrator(ctrl)

	m.EXPECT().Up().Return(nil).MaxTimes(1)
	m.EXPECT().Close().Return(nil, nil).MaxTimes(1)
	m.EXPECT().Up().Return(errors.New("")).MaxTimes(1)
	m.EXPECT().Up().Return(nil).MaxTimes(2)
	m.EXPECT().Close().Return(errors.New(""), nil).MaxTimes(1)
	m.EXPECT().Close().Return(nil, errors.New("")).MaxTimes(1)

	err := ApplyMigrations(m)
	require.NoError(t, err)

	err = ApplyMigrations(m)
	require.Error(t, err)

	err = ApplyMigrations(m)
	require.Error(t, err)

	err = ApplyMigrations(m)
	require.Error(t, err)
}