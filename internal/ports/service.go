package ports

import (
	"context"
	"test_task/internal/entities"
	"test_task/pkg/dto"
)

type Service interface {
	// ========== USER ==========
	CreateUser(ctx context.Context, name, email, password string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)
	Login(ctx context.Context, email, password string) (*entities.User, string, error)

	// ========== TEAM ==========
	CreateTeam(ctx context.Context, userID int64, name string) (*entities.Team, error)
	GetTeams(ctx context.Context, userID int64) ([]dto.TeamWithRole, error)
	GetTeamByID(ctx context.Context, teamID int64) (*entities.Team, error)
	GetTeamMembers(ctx context.Context, teamID int64) ([]*entities.TeamMember, error)

	// ========== TEAM MEMBER ==========
	AddMember(ctx context.Context, userID, teamID, memberID int64, role string) error

	// ========== TASK ==========
	CreateTask(ctx context.Context, userID, assigneeID, teamID int64, title, description string) (*entities.Task, error)
	GetTaskByID(ctx context.Context, taskID int64) (*entities.Task, error)
	GetTasksByTeam(ctx context.Context, teamID int64, limit, offset int) ([]*entities.Task, error)
	UpdateTask(ctx context.Context, userID int64, task *entities.Task) error

	// ========== COMMENT ==========
	AddComment(ctx context.Context, userID, taskID int64, content string) (*entities.TaskComment, error)
	GetCommentsByTask(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskComment, error)

	// ========== HISTORY ==========
	GetTaskHistory(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error)
}
