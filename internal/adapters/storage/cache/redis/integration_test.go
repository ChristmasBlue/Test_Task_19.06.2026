package redis_test

import (
	"context"
	"test_task/internal/entities"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"test_task/internal/adapters/storage/cache/redis"
)

func TestRedisIntegration_SetGetTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// 1. Подключаемся к реальному Redis
	cache, err := redis.NewRedisCache(&testConfig{})
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	defer cache.Stop(ctx)

	// 2. Создаём задачу
	task := &entities.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      entities.TaskStatusTODO,
	}

	// 3. Сохраняем в кеш
	err = cache.SetTask(ctx, task)
	require.NoError(t, err)

	// 4. Получаем из кеша
	found, err := cache.GetTask(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, task.ID, found.ID)
	require.Equal(t, task.Title, found.Title)
	require.Equal(t, task.Status, found.Status)
}
