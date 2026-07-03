package cases_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"test_task/internal/adapters/storage/cache/redis"
	"test_task/internal/adapters/storage/database/ms"
	"test_task/internal/cases"
)

func setupService(t *testing.T) (*cases.Service, *ms.Storage, func()) {
	t.Helper()

	storage, err := ms.NewStorage(&testConfig{})
	require.NoError(t, err)

	cache, err := redis.NewRedisCache(&testConfig{})
	require.NoError(t, err)

	svc, err := cases.NewService(storage, cache, &testConfig{})
	require.NoError(t, err)

	cleanup := func() {
		storage.Stop(context.Background())
		cache.Stop(context.Background())
	}

	return svc, storage, cleanup
}

func TestServiceIntegration_CreateTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	svc, storage, cleanup := setupService(t)
	defer cleanup()

	ctx := context.Background()

	user, err := storage.CreateUser(ctx, "John Doe", "john@example.com", "hashed_password")
	require.NoError(t, err)

	team, err := storage.CreateTeam(ctx, "Team A", user.ID)
	require.NoError(t, err)

	err = storage.AddMember(ctx, user.ID, team.ID, "owner")
	require.NoError(t, err)

	task, err := svc.CreateTask(ctx, user.ID, user.ID, team.ID, "Test Task", "Test Description")
	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, "Test Task", task.Title)

	found, err := storage.GetTaskByID(ctx, task.ID)
	require.NoError(t, err)
	require.Equal(t, task.ID, found.ID)

	cached, err := svc.GetTaskByID(ctx, user.ID, task.ID)
	require.NoError(t, err)
	require.Equal(t, task.ID, cached.ID)
}
