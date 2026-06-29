package cases

import (
	"context"
	"log/slog"
	"test_task/internal/entities"
	"test_task/pkg/dto"
	"test_task/pkg/tool"
	"test_task/tools/config"
	"time"

	"github.com/pkg/errors"
)

type Service struct {
	Storage Repository
	Config  config.Config
}

func NewService(storage Repository, cfg config.Config) (*Service, error) {
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

	if _, err := s.Storage.GetUserByEmail(ctx, email); err == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "email is exist")

	} else if !errors.Is(err, entities.ErrNotFound) {
		slog.Error("GetUserByEmail", "err", err)
		return nil, errors.Wrap(entities.ErrInvalidParam, "failed to check user existence")
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

func (s *Service) AddMember(ctx context.Context, userID, teamID, memberID int64, role string) error {
	slog.Info("AddMember")
	if memberID <= 0 {
		slog.Error("AddMember", "err", "invalid memberID")
		return errors.Wrap(entities.ErrInvalidParam, "memberID is invalid")
	}

	if teamID <= 0 {
		slog.Error("AddMember", "err", "invalid teamID")
		return errors.Wrap(entities.ErrInvalidParam, "teamID is invalid")
	}

	if role == "" {
		role = "member"
	}

	_, err := s.Storage.GetTeamByID(ctx, teamID)
	if err != nil {
		slog.Error("GetTeam", "err", err)
		return errors.Wrap(err, "get team")
	}

	isAdmin, err := s.Storage.IsAdminOrOwner(ctx, userID, teamID)
	if err != nil {
		slog.Error("IsAdminOrOwner", "err", err)
		return errors.Wrap(err, "failed to check user permissions")
	}

	if !isAdmin {
		return errors.Wrap(entities.ErrInvalidParam, "user is not an admin")
	}

	_, err = s.Storage.GetUserByID(ctx, memberID)
	if err != nil {
		slog.Error("GetUserByID", "err", err)
		return errors.Wrap(err, "failed to get member")
	}

	err = s.Storage.AddMember(ctx, memberID, teamID, role)

	if err != nil {
		slog.Error("AddMember", "err", err)
		return errors.Wrap(err, "failed to add member")
	}

	return nil
}

func (s *Service) CreateTeam(ctx context.Context, userID int64, name string) (*entities.Team, error) {
	slog.Info("CreateTeam")
	if name == "" {
		slog.Error("CreateTeam", "err", "empty name")
		return nil, errors.Wrap(entities.ErrInvalidParam, "name is empty")
	}

	var team *entities.Team
	var err error

	err = s.Storage.InTx(ctx, func(ctx context.Context, repo Repository) error {
		team, err = repo.CreateTeam(ctx, name, userID)
		if err != nil {
			slog.Error("CreateTeam", "err", err)
			return errors.Wrap(err, "create team failed")
		}

		err = repo.AddMember(ctx, userID, team.GetID(), "owner")
		if err != nil {
			slog.Error("AddMember", "err", err)
			return errors.Wrap(err, "add member failed")
		}

		return nil
	})

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

func (s *Service) GetTeams(ctx context.Context, userID int64) ([]dto.TeamWithRole, error) {
	slog.Info("GetTeams")

	teamIDs, err := s.Storage.GetUserTeams(ctx, userID)
	if err != nil {
		slog.Error("GetUserTeams", "err", err)
		return nil, errors.Wrap(err, "failed to get user teams")
	}

	if len(teamIDs) == 0 {
		return []dto.TeamWithRole{}, nil
	}

	teams, err := s.Storage.GetTeamsByIDs(ctx, teamIDs)
	if err != nil {
		slog.Error("GetTeamsByIDs", "err", err)
		return nil, errors.Wrap(err, "failed to get teams info")
	}

	result := make([]dto.TeamWithRole, 0, len(teams))
	for _, team := range teams {
		role, err := s.Storage.GetMemberRole(ctx, userID, team.ID)
		if err != nil {
			slog.Error("GetMemberRole", "err", err, "team_id", team.ID)
			return nil, errors.Wrap(err, "failed to get member role")
		}

		result = append(result, dto.TeamWithRole{
			Team: team,
			Role: role,
		})
	}

	return result, nil
}

func (s *Service) GetTeamByID(ctx context.Context, teamID int64) (*entities.Team, error) {
	slog.Info("GetTeamByID")
	team, err := s.Storage.GetTeamByID(ctx, teamID)
	if err != nil {
		slog.Error("GetTeamByID", "err", err)
		return nil, errors.Wrap(err, "get team failed")
	}

	return team, nil
}

func (s *Service) GetTeamMembers(ctx context.Context, teamID int64) ([]*entities.TeamMember, error) {
	slog.Info("GetTeamMembers")
	members, err := s.Storage.GetTeamMembers(ctx, teamID)
	if err != nil {
		slog.Error("GetTeamMembers", "err", err)
		return nil, errors.Wrap(err, "get team members failed")
	}
	return members, nil
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

	var task *entities.Task
	var err error
	status := entities.TaskStatusTODO

	err = s.Storage.InTx(ctx, func(ctx context.Context, repo Repository) error {
		task, err = repo.CreateTask(ctx, userID, assigneeID, teamID, title, description, status)
		if err != nil {
			slog.Error("CreateTask", "err", err)
			return errors.Wrap(err, "create task failed")
		}

		newHistory, err := entities.NewTaskHistory(task.GetID(), userID, "created", nil, &title, time.Now())
		if err != nil {
			slog.Error("CreateTask", "err", err)
			return errors.Wrap(err, "create task history failed")
		}

		err = repo.AddHistoryRecord(ctx, newHistory)
		if err != nil {
			slog.Error("CreateTask", "err", err)
			return errors.Wrap(err, "add history record failed")
		}

		return nil
	})

	if err != nil {
		slog.Error("CreateTask", "err", err)
		return nil, errors.Wrap(err, "create task failed")
	}

	return task, nil
}

func (s *Service) GetTaskByID(ctx context.Context, taskID int64) (*entities.Task, error) {
	slog.Info("GetTaskByID")
	if taskID <= 0 {
		slog.Error("GetTaskByID", "err", "invalid taskID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "taskID is invalid")
	}

	task, err := s.Storage.GetTaskByID(ctx, taskID)
	if err != nil {
		slog.Error("GetTaskByID", "err", err)
		return nil, errors.Wrap(err, "get task failed")
	}

	return task, nil
}

func (s *Service) GetTasksByTeam(ctx context.Context, teamID int64, limit, offset int) ([]*entities.Task, error) {
	slog.Info("GetTasksByTeam")
	tasks, err := s.Storage.GetTasksByTeam(ctx, teamID, limit, offset)
	if err != nil {
		slog.Error("GetTasksByTeam", "err", err)
		return nil, errors.Wrap(err, "get tasks failed")
	}
	return tasks, nil
}

func (s *Service) UpdateTask(ctx context.Context, userID int64, task *entities.Task) error {
	slog.Info("UpdateTask")

	if task == nil {
		slog.Error("UpdateTask", "err", "invalid task")
		return errors.Wrap(entities.ErrInvalidParam, "task is empty")
	}
	if userID <= 0 {
		return errors.Wrap(entities.ErrInvalidParam, "userID is invalid")
	}

	oldTask, err := s.Storage.GetTaskByID(ctx, task.GetID())
	if err != nil {
		slog.Error("GetTaskByID", "err", err)
		return errors.Wrap(err, "failed to get task")
	}

	slog.Info("Assignee check",
		"old_assignee", oldTask.AssigneeID,
		"new_assignee", task.AssigneeID,
	)

	if task.AssigneeID == 0 {
		task.AssigneeID = oldTask.AssigneeID
	}

	if task.Title == "" {
		task.Title = oldTask.Title
	}

	if task.Description == "" {
		task.Description = oldTask.Description
	}

	if task.Status == "" {
		task.Status = oldTask.Status
	}
	task.TeamID = oldTask.TeamID
	task.ID = oldTask.ID
	task.OwnerID = oldTask.OwnerID
	task.CreateAt = oldTask.CreateAt

	var historyRecords []*entities.TaskHistory

	if oldTask.Title != task.Title {
		record, err := entities.NewTaskHistory(
			task.GetID(),
			userID,
			"title",
			&oldTask.Title,
			&task.Title,
			time.Now(),
		)
		if err != nil {
			return errors.Wrap(err, "create history record failed")
		}
		historyRecords = append(historyRecords, record)
	}

	if oldTask.Description != task.Description {
		record, err := entities.NewTaskHistory(
			task.GetID(),
			userID,
			"description",
			&oldTask.Description,
			&task.Description,
			time.Now(),
		)
		if err != nil {
			return errors.Wrap(err, "create history record failed")
		}
		historyRecords = append(historyRecords, record)
	}

	if oldTask.Status != task.Status {
		record, err := entities.NewTaskHistory(
			task.GetID(),
			userID,
			"status",
			&oldTask.Status,
			&task.Status,
			time.Now(),
		)
		if err != nil {
			return errors.Wrap(err, "create history record failed")
		}
		historyRecords = append(historyRecords, record)
	}

	if len(historyRecords) == 0 {
		slog.Info("No changes detected", "task_id", task.GetID())
		return nil
	}

	err = s.Storage.InTx(ctx, func(ctx context.Context, repo Repository) error {
		err := repo.UpdateTask(ctx, task)
		if err != nil {
			slog.Error("UpdateTask", "err", err)
			return errors.Wrap(err, "update task failed")
		}

		for _, record := range historyRecords {
			err := repo.AddHistoryRecord(ctx, record)
			if err != nil {
				slog.Error("AddHistoryRecord", "err", err)
				return errors.Wrap(err, "add history record failed")
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("UpdateTask", "err", err)
		return err
	}

	return nil

}

func (s *Service) AddComment(ctx context.Context, userID, taskID int64, content string) (*entities.TaskComment, error) {
	slog.Info("AddComment")

	if userID <= 0 {
		slog.Error("AddComment", "err", "invalid userID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "userID is invalid")
	}
	if taskID <= 0 {
		slog.Error("AddComment", "err", "invalid taskID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "taskID is invalid")
	}
	if content == "" {
		slog.Error("AddComment", "err", "empty content")
		return nil, errors.Wrap(entities.ErrInvalidParam, "content is empty")
	}

	task, err := s.Storage.GetTaskByID(ctx, taskID)
	if err != nil {
		slog.Error("GetTaskByID", "err", err)
		return nil, errors.Wrap(err, "failed to get task")
	}

	isMember, err := s.Storage.IsMember(ctx, userID, task.TeamID)
	if err != nil {
		slog.Error("IsMember", "err", err)
		return nil, errors.Wrap(err, "failed to check membership")
	}
	if !isMember {
		return nil, errors.Wrap(entities.ErrInvalidParam, "user is not a member of the team")
	}

	comment, err := entities.NewTaskComment(taskID, userID, content, time.Now())
	if err != nil {
		slog.Error("NewTaskComment", "err", err)
		return nil, errors.Wrap(err, "create comment failed")
	}

	err = s.Storage.AddComment(ctx, comment)
	if err != nil {
		slog.Error("AddComment", "err", err)
		return nil, errors.Wrap(err, "add comment failed")
	}

	return comment, nil

}

func (s *Service) GetCommentsByTask(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskComment, error) {
	slog.Info("GetCommentsByTask")
	if taskID <= 0 {
		slog.Error("GetCommentsByTask", "err", "invalid taskID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "taskID is invalid")
	}

	comms, err := s.Storage.GetCommentsByTask(ctx, taskID, limit, offset)
	if err != nil {
		slog.Error("GetCommentsByTask", "err", err)
		return nil, errors.Wrap(err, "get comments failed")
	}

	return comms, nil
}

func (s *Service) GetTaskHistory(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error) {
	slog.Info("GetHistory")
	if taskID <= 0 {
		slog.Error("GetHistory", "err", "invalid taskID")
		return nil, errors.Wrap(entities.ErrInvalidParam, "taskID is invalid")
	}

	history, err := s.Storage.GetTaskHistory(ctx, taskID, limit, offset)
	if err != nil {
		slog.Error("GetHistory", "err", err)
		return nil, errors.Wrap(err, "get history failed")
	}

	return history, nil
}

//GetTasksByFilters() error
