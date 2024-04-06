package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	token, err := CreateJWT(false, []byte(""), time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	_, err = CheckIsAdminInJWT(token, "abc")
	require.Error(t, err)

	isAdmin, err := CheckIsAdminInJWT(token, "")
	require.NoError(t, err)
	require.Equal(t, false, isAdmin)

	token, err = CreateJWT(true, []byte(""), time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	isAdmin, err = CheckIsAdminInJWT(token, "")
	require.NoError(t, err)
	require.Equal(t, true, isAdmin)

	expiredStr, err := CreateJWT(false, []byte(""), time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	_, err = CheckIsAdminInJWT(expiredStr, "")
	require.Error(t, err)
}
