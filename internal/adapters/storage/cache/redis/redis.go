package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"test_task/internal/entities"
	"test_task/tools/config"
	"time"

	"github.com/pkg/errors"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

var (
	taskKey = "task:%d"
	teamKey = "team:tasks:%d"
)

func NewRedisCache(cfg config.Config) (*RedisCache, error) {
	addr := fmt.Sprintf("%s:%d", cfg.CacheHost(), cfg.CachePort())
	slog.Info("Redis address", "addr", addr)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.CachePass(),
		DB:       cfg.CacheDB(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		slog.Error("Connection to redis", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, "failed to connect to redis")
	}

	return &RedisCache{client: client, ttl: cfg.CacheTTL()}, nil
}

func (c *RedisCache) Start() error {
	return nil
}

func (c *RedisCache) Stop(ctx context.Context) error {
	<-ctx.Done()
	slog.Info("Closing Redis connection")
	return c.client.Close()
}

func (c *RedisCache) GetTask(ctx context.Context, taskID int64) (*entities.Task, error) {
	slog.Debug("GetTask")

	key := fmt.Sprintf(taskKey, taskID)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		slog.Error("GetTask", "err", err)
		return nil, errors.Wrap(err, "failed to get task from redis")
	}

	var task *entities.Task
	if err := json.Unmarshal([]byte(data), task); err != nil {
		slog.Error("GetTask", "err", err)
		return nil, errors.Wrap(err, "failed to unmarshal task")
	}

	return task, nil
}

func (c *RedisCache) GetTasksByTeam(ctx context.Context, teamID int64) ([]*entities.Task, error) {
	slog.Debug("GetTasksByTeam")

	key := fmt.Sprintf(teamKey, teamID)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		slog.Error("GetTasksByTeam", "err", err)
		return nil, errors.Wrap(err, "failed to get team tasks from redis")
	}

	var tasks []*entities.Task
	if err := json.Unmarshal([]byte(data), &tasks); err != nil {
		slog.Error("GetTasksByTeam", "err", err)
		return nil, errors.Wrap(err, "failed to unmarshal tasks")
	}

	return tasks, nil
}

func (c *RedisCache) SetTask(ctx context.Context, task *entities.Task) error {
	slog.Debug("SetTask")

	key := fmt.Sprintf(taskKey, task.ID)

	data, err := json.Marshal(task)
	if err != nil {
		slog.Error("SetTask", "err", err)
		return errors.Wrap(err, "failed to marshal task")
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		slog.Error("SetTask", "err", err)
		return errors.Wrap(err, "failed to set task to redis")
	}

	return nil
}

func (c *RedisCache) SetTasksByTeam(ctx context.Context, teamID int64, tasks []*entities.Task) error {
	slog.Debug("SetTasksByTeam")

	key := fmt.Sprintf(teamKey, teamID)

	data, err := json.Marshal(tasks)
	if err != nil {
		slog.Error("SetTasksByTeam", "err", err)
		return errors.Wrap(err, "failed to marshal tasks")
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		slog.Error("SetTasksByTeam", "err", err)
		return errors.Wrap(err, "failed to set team tasks to redis")
	}

	return nil
}

func (c *RedisCache) DeleteTask(ctx context.Context, taskID int64) error {
	slog.Debug("DeleteTask")

	key := fmt.Sprintf(taskKey, taskID)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		slog.Error("DeleteTask", "err", err)
		return errors.Wrap(err, "failed to delete task from redis")
	}

	return nil
}

func (c *RedisCache) DeleteTasksByTeam(ctx context.Context, teamID int64) error {
	slog.Debug("DeleteTasksByTeam")

	key := fmt.Sprintf(teamKey, teamID)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		slog.Error("DeleteTasksByTeam", "err", err)
		return errors.Wrap(err, "failed to delete team tasks from redis")
	}

	slog.Debug("DeleteTasksByTeam", "deleted", key)
	return nil
}
