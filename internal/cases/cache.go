package cases

//go:generate go run go.uber.org/mock/mockgen@latest -source=cache.go -destination=mocks/cache_mock.go -package=mocks Cache
import (
	"context"
	"test_task/internal/entities"
	"test_task/pkg/dto"
)

type Cache interface {
	Start() error

	Stop(ctx context.Context) error

	GetTask(ctx context.Context, taskID int64) (*entities.Task, error)

	GetTasksByTeam(ctx context.Context, teamID int64) ([]*entities.Task, error)

	GetTasksByFilter(ctx context.Context, filter dto.TaskFilter) ([]*entities.Task, error)

	SetTasksByFilter(ctx context.Context, filter dto.TaskFilter, tasks []*entities.Task) error

	SetTask(ctx context.Context, task *entities.Task) error

	SetTasksByTeam(ctx context.Context, teamID int64, tasks []*entities.Task) error

	DeleteTask(ctx context.Context, taskID int64) error

	DeleteTasksByTeam(ctx context.Context, teamID int64) error

	DeleteTasksByFilter(ctx context.Context, taskID, teamID int64) error
}
