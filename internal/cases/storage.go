package cases

import (
	"context"
	"test_task/internal/entities"
)

type Storage interface {
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUserByID(ctx context.Context, id int64) (*entities.User, error)
	CreateUser(ctx context.Context, name, email, hashPass string) (*entities.User, error)
	//AddTeamToUser(ctx context.Context, user *entities.User, team *entities.Team) error

	GetTeam(ctx context.Context, teamID int64) (*entities.Team, error)
	CreateTeam(ctx context.Context, userID int64, name string) (*entities.Team, error)
	GetAllTeams(ctx context.Context, userID int64) ([]*entities.Team, error)
	//AddTaskTeam(ctx context.Context, task *entities.Task, team *entities.Team) error
	AddMemberInTeam(ctx context.Context, userID, teamID int64) error

	GetTask(ctx context.Context, taskID int64) (*entities.Task, error)
	CreateTask(ctx context.Context, user, assigneeID, teamID int64, title, description string) (*entities.Task, error)
	UpdateTask(ctx context.Context, task *entities.Task) error
	AddComment(ctx context.Context, taskID int64, comment *entities.Comment) error

	//GetTaskByFilters(ctx context.Context) ([]*entities.Task, error)

	GetHistory(ctx context.Context, taskID int64) (*entities.History, error)
	AddHistory(ctx context.Context, record *entities.Record) error
}
