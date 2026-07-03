package cases

//go:generate go run go.uber.org/mock/mockgen@latest -source=storage.go -destination=mocks/repository_mock.go -package=mocks Repository
import (
	"context"
	"test_task/internal/entities"
	"test_task/pkg/dto"
)

// Repository — основной интерфейс репозитория
type Repository interface {

	// InTx выполняет функцию в транзакции
	InTx(ctx context.Context, fn func(ctx context.Context, repo Repository) error) error

	Start() error
	Stop(ctx context.Context) error
	Ping(ctx context.Context) error

	// ========== USER ==========
	CreateUser(ctx context.Context, name, email, hashPass string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)

	// ========== TEAM ==========
	CreateTeam(ctx context.Context, name string, ownerID int64) (*entities.Team, error)
	GetTeamByID(ctx context.Context, id int64) (*entities.Team, error)
	GetUserTeams(ctx context.Context, userID int64) ([]int64, error)
	GetTeamsByIDs(ctx context.Context, ids []int64) ([]*entities.Team, error)

	// ========== TEAM MEMBER ==========
	AddMember(ctx context.Context, userID, teamID int64, role string) error
	GetTeamMembers(ctx context.Context, teamID int64) ([]*entities.TeamMember, error)
	IsAdminOrOwner(ctx context.Context, userID, teamID int64) (bool, error)
	IsMember(ctx context.Context, userID, teamID int64) (bool, error)
	GetMemberRole(ctx context.Context, userID, teamID int64) (string, error)

	// ========== TASK ==========
	CreateTask(ctx context.Context, userID, assigneeID, teamID int64, title, description, status string) (*entities.Task, error)
	GetTaskByID(ctx context.Context, id int64) (*entities.Task, error)
	GetTasksByTeam(ctx context.Context, teamID int64, limit, offset int) ([]*entities.Task, error)
	UpdateTask(ctx context.Context, task *entities.Task) error
	GetTasksByFilter(ctx context.Context, filter dto.TaskFilter) ([]*entities.Task, error)

	// ========== TASK HISTORY ==========
	AddHistoryRecord(ctx context.Context, record *entities.TaskHistory) error
	GetTaskHistory(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error)

	// ========== TASK COMMENT ==========
	AddComment(ctx context.Context, comment *entities.TaskComment) error
	GetCommentsByTask(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskComment, error)

	GetInvalidAssigneeTasks(ctx context.Context) ([]dto.InvalidAssigneeTask, error)
	GetTeamStats(ctx context.Context) ([]dto.TeamStats, error)
	GetTopCreators(ctx context.Context) ([]dto.TopCreator, error)
}
