package ports

import (
	"context"
	"test_task/internal/entities"
)

type Service interface {
	CreateUser(ctx context.Context, name, email, password string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)
	Login(ctx context.Context, email, password string) (*entities.User, string, error)

	CreateTeam(ctx context.Context, userID int64, name string) (*entities.Team, error)
	AddMember(ctx context.Context, userID, teamID, memberID int64) error
	GetTeams(ctx context.Context, userID int64) ([]*entities.Team, error)
	GetTeamByID(ctx context.Context, userID, teamID int64) (*entities.Team, error)

	CreateTask(ctx context.Context, user, assigneeID, teamID int64, title, description string) (*entities.Task, error)
	GetTaskByID(ctx context.Context, taskID int64) (*entities.Task, error)
	//GetTasksByFilters() error
	UpdateTask(ctx context.Context, userID int64, task *entities.Task) error
	AddComment(ctx context.Context, userID, taskID int64, comment string) error

	GetHistory(ctx context.Context, taskID int64) (*entities.History, error)
}
