package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"test_task/internal/entities"
	"test_task/pkg/dto"
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

func (c *RedisCache) GetTasksByFilter(ctx context.Context, filter dto.TaskFilter) ([]*entities.Task, error) {
	key := c.buildFilterKey(filter)
	slog.Debug("GetTasksByFilter", "key", key)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		slog.Error("GetTasksByFilter", "err", err)
		return nil, errors.Wrap(err, "failed to get tasks from redis")
	}

	var tasks []*entities.Task
	if err := json.Unmarshal([]byte(data), &tasks); err != nil {
		slog.Error("GetTasksByFilter", "err", err)
		return nil, errors.Wrap(err, "failed to unmarshal tasks")
	}

	return tasks, nil
}

func (c *RedisCache) SetTasksByFilter(ctx context.Context, filter dto.TaskFilter, tasks []*entities.Task) error {
	key := c.buildFilterKey(filter)
	slog.Debug("SetTasksByFilter", "key", key, "count", len(tasks))

	if len(tasks) == 0 {
		return nil
	}

	data, err := json.Marshal(tasks)
	if err != nil {
		slog.Error("SetTasksByFilter", "err", err)
		return errors.Wrap(err, "failed to marshal tasks")
	}

	if err := c.client.Set(ctx, key, data, c.ttl).Err(); err != nil {
		slog.Error("SetTasksByFilter", "err", err)
		return errors.Wrap(err, "failed to set tasks to redis")
	}

	pipe := c.client.Pipeline()
	for _, task := range tasks {
		indexKey := c.filterIndexKey(task.ID)
		pipe.SAdd(ctx, indexKey, key)
		pipe.Expire(ctx, indexKey, c.ttl)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		slog.Warn("Failed to create indexes", "err", err)
	}

	slog.Debug("SetTasksByFilter", "indexed_tasks", len(tasks))
	return nil
}

func (c *RedisCache) DeleteTasksByFilter(ctx context.Context, taskID, teamID int64) error {
	indexKey := c.filterIndexKey(taskID)
	slog.Debug("DeleteTasksByFilter", "task_id", taskID, "team_id", teamID)

	filterKeys, err := c.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		slog.Error("DeleteTasksByFilter", "err", err)
		return errors.Wrap(err, "failed to get filter keys")
	}

	if err := c.client.Del(ctx, filterKeys...).Err(); err != nil {
		slog.Error("DeleteTasksByFilter", "err", err)
		return errors.Wrap(err, "failed to delete filter keys")
	}

	if err := c.client.Del(ctx, indexKey).Err(); err != nil {
		slog.Warn("Failed to delete index", "err", err)
	}

	slog.Debug("DeleteTasksByFilter", "deleted", len(filterKeys), "task_id", taskID)
	return nil
}

func (c *RedisCache) buildFilterKey(filter dto.TaskFilter) string {
	teamIDs := make([]int64, len(filter.TeamIDs))
	copy(teamIDs, filter.TeamIDs)
	if len(teamIDs) > 1 {
		sort.Slice(teamIDs, func(i, j int) bool { return teamIDs[i] < teamIDs[j] })
	}

	parts := make([]string, len(teamIDs))
	for i, id := range teamIDs {
		parts[i] = strconv.FormatInt(id, 10)
	}
	teamIDsStr := strings.Join(parts, ",")

	status := "none"
	if filter.Status != nil && *filter.Status != "" {
		status = *filter.Status
	}

	assignee := "none"
	if filter.AssigneeID != nil {
		assignee = strconv.FormatInt(*filter.AssigneeID, 10)
	}

	return fmt.Sprintf("tasks:filter:team:%s:status:%s:assignee:%s:limit:%d:offset:%d",
		teamIDsStr, status, assignee, filter.Limit, filter.Offset)
}

func (c *RedisCache) filterIndexKey(taskID int64) string {
	return fmt.Sprintf("task:filters:%d", taskID)
}
