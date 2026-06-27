package cases

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"test_task/internal/entities"
	"test_task/pkg/tool"
	"test_task/tools/config"

	"github.com/pkg/errors"
)

type Service struct {
	Storage Storage
	Config  config.Config
}

func NewService(storage Storage, cfg config.Config) (*Service, error) {
	if storage == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "storage not set")
	}

	if cfg == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "config not set")
	}

	return &Service{
		Storage: storage,
		Config:  cfg,
	}, nil
}

func (s *Service) CreateUser(ctx context.Context, name, email, password string) (*entities.User, error) {
	slog.Info("CreateUser")
	if name == "" {
		slog.Error("CreateUser", "err", "empty name")
		return nil, errors.Wrap(entities.ErrInvalidParam, "name is empty")
	}

	if email == "" {
		slog.Error("CreateUser", "err", "empty email")
		return nil, errors.Wrap(entities.ErrInvalidParam, "email is empty")
	}

	if password == "" {
		slog.Error("CreateUser", "err", "empty password")
		return nil, errors.Wrap(entities.ErrInvalidParam, "password is empty")
	}

	if _, err := s.Storage.GetUserByEmail(ctx, email); err != nil {
		slog.Error("GetUserByEmail", "err", err)
		return nil, errors.Wrap(err, "email is exist")
	}

	hashPass, err := tool.HashingPass(password)
	if err != nil {
		slog.Error("HashingPass", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	user, err := s.Storage.CreateUser(ctx, name, email, string(hashPass))
	if err != nil {
		slog.Error("CreateUser", "err", err)
		return nil, errors.Wrap(err, "create user")
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*entities.User, string, error) {
	slog.Info("Login")
	user, err := s.Storage.GetUserByEmail(ctx, email)
	if err != nil {
		slog.Error("GetUserByEmail", "err", err)
		return nil, "", entities.ErrInvalidParam
	}

	ok, err := tool.CheckHashPass(user.HashPassword, []byte(password))
	if err != nil {
		slog.Error("CheckHashPass", "err", err)
		return nil, "", errors.Wrap(entities.ErrInternal, "check hash")
	}
	if !ok {
		slog.Error("CheckHashPass", "err", "hash is different")
		return nil, "", entities.ErrInvalidParam
	}

	token, err := tool.GenerateJWTToken(user.ID, s.Config.TokenDuration(), []byte(s.Config.TokenSecret()))
	if err != nil {
		slog.Error("GenerateJWTToken", "err", err)
		return nil, "", errors.Wrap(entities.ErrInternal, "generate token")
	}
	return user, token, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	slog.Info("GetUserByEmail")
	if email == "" {
		slog.Error("GetUserByEmail", "err", "empty email")
		return nil, errors.Wrap(entities.ErrInvalidParam, "email is empty")
	}

	user, err := s.Storage.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return nil, errors.Wrap(entities.ErrNotFound, "user not found")
		}
		slog.Error("GetUserByEmail", "err", err)
		return nil, errors.Wrap(err, "get user by email")
	}

	return user, nil
}

func (s *Service) GetUserByID(ctx context.Context, id int64) (*entities.User, error) {
	slog.Info("GetUserByID")
	if id <= 0 {
		slog.Error("GetUserByID", "err", "invalid id")
		return nil, errors.Wrap(entities.ErrInvalidParam, "invalid id")
	}

	user, err := s.Storage.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, entities.ErrNotFound) {
			return nil, errors.Wrap(entities.ErrNotFound, "user not found")
		}
		slog.Error("GetUserByID", "err", err)
		return nil, errors.Wrap(err, "get user by id")
	}
	return user, nil
}

func (s *Service) AddMember(ctx context.Context, userID int64, teamID, memberID int64) error {
	slog.Info("AddMember")
	if memberID <= 0 {
		slog.Error("AddMember", "err", "invalid memberID")
		return errors.Wrap(entities.ErrInvalidParam, "memberID is invalid")
	}

	if teamID <= 0 {
		slog.Error("AddMember", "err", "invalid teamID")
		return errors.Wrap(entities.ErrInvalidParam, "teamID is invalid")
	}

	team, err := s.Storage.GetTeam(ctx, teamID)
	if err != nil {
		slog.Error("GetTeam", "err", err)
		return errors.Wrap(err, "get team")
	}

	if team.OwnerID != userID {
		return errors.Wrap(entities.ErrInvalidParam, "inviterID is not owner of team")
	}

	member, err := s.Storage.GetUserByID(ctx, memberID)
	if err != nil {
		slog.Error("GetUserByID", "err", err)
		return errors.Wrap(err, "failed to get member")
	}

	if !slices.Contains(team.MembersID, memberID) || !slices.Contains(member.TeamsID, team.ID) {
		err = s.Storage.AddMemberInTeam(ctx, memberID, teamID)
		if err != nil {
			slog.Error("AddMember", "err", err)
			return errors.Wrap(err, "add member team failed")
		}

		err = team.AddMemberID(member.ID)
		if err != nil {
			slog.Error("AddMember", "err", err)
			return errors.Wrap(err, "add member failed")
		}

		err = member.AddTeamID(team.ID)
		if err != nil {
			slog.Error("AddMember", "err", err)
			return errors.Wrap(err, "add member failed")
		}
	}

	return nil
}

func (s *Service) CreateTeam(ctx context.Context, userID int64, name string) (*entities.Team, error) {
	slog.Info("CreateTeam")
	if name == "" {
		slog.Error("CreateTeam", "err", "empty name")
		return nil, errors.Wrap(entities.ErrInvalidParam, "name is empty")
	}

	team, err := s.Storage.CreateTeam(ctx, userID, name)
	if err != nil {
		slog.Error("CreateTeam", "err", err)
		return nil, errors.Wrap(err, "create team failed")
	}

	err = team.AddMemberID(userID)
	if err != nil {
		slog.Error("AddMember", "err", err)
		return nil, errors.Wrap(err, "add member failed")
	}

	return team, nil
}

func (s *Service) GetTeams(ctx context.Context, userID int64) ([]*entities.Team, error) {
	slog.Info("GetTeams")
	teams, err := s.Storage.GetAllTeams(ctx, userID)
	if err != nil {
		slog.Error("GetTeams", "err", err)
		return nil, errors.Wrap(err, "get all teams failed")
	}

	return teams, nil
}

func (s *Service) GetTeamByID(ctx context.Context, userID, teamID int64) (*entities.Team, error) {
	slog.Info("GetTeamByID")
	team, err := s.Storage.GetTeam(ctx, teamID)
	if err != nil {
		slog.Error("GetTeamByID", "err", err)
		return nil, errors.Wrap(err, "get team failed")
	}

	if !slices.Contains(team.MembersID, userID) {
		slog.Error("GetTeamByID", "err", "member not found in team")
		return nil, errors.Wrap(entities.ErrInvalidParam, "user is not member of team")
	}
	return team, nil
}

func (s *Service) CreateTask(ctx context.Context, userID, assigneeID, teamID int64, title, description string) (*entities.Task, error) {
	slog.Info("CreateTask")
	if title == "" {
		slog.Error("CreateTask", "err", "empty title")
		return nil, errors.Wrap(entities.ErrInvalidParam, "title is empty")
	}
	if description == "" {
		slog.Error("CreateTask", "err", "empty description")
		return nil, errors.Wrap(entities.ErrInvalidParam, "description is empty")
	}
	if assigneeID <= 0 {
		slog.Error("CreateTask", "err", "invalid assigneeID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "assigneeID is invalid")
	}
	if teamID <= 0 {
		slog.Error("CreateTask", "err", "invalid teamID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "teamID is invalid")
	}

	team, err := s.Storage.GetTeam(ctx, teamID)
	if err != nil {
		slog.Error("GetTeam", "err", err)
		return nil, errors.Wrap(err, "get team failed")
	}

	task, err := s.Storage.CreateTask(ctx, userID, assigneeID, teamID, title, description)
	if err != nil {
		slog.Error("CreateTask", "err", err)
		return nil, errors.Wrap(err, "create task failed")
	}

	newRecord, err := entities.NewRecord(userID, task.ID, entities.StatusCreated, "task created", "", task.CreateAt)
	if err != nil {
		slog.Error("NewRecord", "err", err)
		return nil, errors.Wrap(err, "create history task failed")
	}

	err = s.Storage.AddHistory(ctx, newRecord)
	if err != nil {
		slog.Error("AddHistory", "err", err)
		return nil, errors.Wrap(err, "add history task failed")
	}

	err = team.AddTaskID(task.ID)
	if err != nil {
		slog.Error("AddTaskID", "err", err)
		return nil, errors.Wrap(err, "add task failed")
	}

	return task, nil
}

func (s *Service) GetTaskByID(ctx context.Context, taskID int64) (*entities.Task, error) {
	slog.Info("GetTaskByID")
	if taskID <= 0 {
		slog.Error("GetTaskByID", "err", "invalid taskID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "taskID is invalid")
	}

	task, err := s.Storage.GetTask(ctx, taskID)
	if err != nil {
		slog.Error("GetTaskByID", "err", err)
		return nil, errors.Wrap(err, "get task failed")
	}

	return task, nil
}

func (s *Service) UpdateTask(ctx context.Context, userID int64, task *entities.Task) error {
	slog.Info("UpdateTask")
	if task == nil {
		slog.Error("UpdateTask", "err", "invalid task")
		return errors.Wrap(entities.ErrInvalidParam, "task is empty")
	}

	if userID != task.OwnerID {
		team, err := s.Storage.GetTeam(ctx, task.TeamID)
		if err != nil {
			slog.Error("GetTeam", "err", err)
			return errors.Wrap(err, "get team failed")
		}

		if team.OwnerID != userID {
			return errors.Wrap(entities.ErrInvalidParam, "user can't update task")
		}
	}

	err := s.Storage.UpdateTask(ctx, task)
	if err != nil {
		slog.Error("UpdateTask", "err", err)
		return errors.Wrap(err, "update task failed")
	}

	newRecord, err := entities.NewRecord(userID, task.ID, entities.StatusUpdated, "task updated", fmt.Sprintf("user id: %d\nstatus: %s\ntitle: %s\ndescription: %s", userID, task.Status, task.Title, task.Description), task.CreateAt)
	if err != nil {
		slog.Error("NewRecord", "err", err)
		return errors.Wrap(err, "update history task failed")
	}

	err = s.Storage.AddHistory(ctx, newRecord)
	if err != nil {
		slog.Error("AddHistory", "err", err)
		return errors.Wrap(err, "add history task failed")
	}

	return nil
}

func (s *Service) AddComment(ctx context.Context, userID, taskID int64, comment string) error {
	slog.Info("AddComment")
	comm, err := entities.NewComment(userID, comment)
	if err != nil {
		slog.Error("AddComment", "err", err)
		return errors.Wrap(err, "create comment failed")
	}

	err = s.Storage.AddComment(ctx, taskID, comm)
	if err != nil {
		slog.Error("AddComment", "err", err)
		return errors.Wrap(err, "add comment failed")
	}

	task, err := s.GetTaskByID(ctx, taskID)
	if err != nil {
		slog.Error("GetTaskByID", "err", err)
		return errors.Wrap(err, "get task failed")
	}

	err = task.AddComment(comm)
	if err != nil {
		slog.Error("AddComment", "err", err)
		return errors.Wrap(err, "add comment failed")
	}

	return nil
}

func (s *Service) GetHistory(ctx context.Context, taskID int64) (*entities.History, error) {
	slog.Info("GetHistory")
	if taskID <= 0 {
		slog.Error("GetHistory", "err", "invalid taskID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "taskID is invalid")
	}

	history, err := s.Storage.GetHistory(ctx, taskID)
	if err != nil {
		slog.Error("GetHistory", "err", err)
		return nil, errors.Wrap(err, "get history failed")
	}

	return history, nil
}

//GetTasksByFilters() error
