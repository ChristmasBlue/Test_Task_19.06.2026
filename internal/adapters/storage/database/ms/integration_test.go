package ms_test

import (
	"context"
	"testing"
	"time"

	"test_task/internal/adapters/storage/database/ms"

	"github.com/stretchr/testify/require"
)

func TestStorageIntegration_CreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := ms.NewStorage(&testConfig{})
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	defer storage.Stop(ctx)

	user, err := storage.CreateUser(ctx, "John Doe", "john@example.com", "hashed_password")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "John Doe", user.Name)
	require.Equal(t, "john@example.com", user.Email)

	found, err := storage.GetUserByEmail(ctx, "john@example.com")
	require.NoError(t, err)
	require.Equal(t, user.ID, found.ID)
}

func TestStorageIntegration_CreateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	storage, err := ms.NewStorage(&testConfig{})
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	defer storage.Stop(ctx)

	user, err := storage.CreateUser(ctx, "Jane Doe", "jane@example.com", "hashed_password")
	require.NoError(t, err)

	team, err := storage.CreateTeam(ctx, "Team A", user.ID)
	require.NoError(t, err)
	require.NotNil(t, team)
	require.Equal(t, "Team A", team.Name)
	require.Equal(t, user.ID, team.OwnerID)
}
