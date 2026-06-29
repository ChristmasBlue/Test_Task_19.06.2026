package public

import (
	"log/slog"
	"net/http"
	"strconv"
	"test_task/internal/entities"

	"github.com/pkg/errors"

	"test_task/pkg/dto"

	"github.com/gin-gonic/gin"
)

func (s *Server) Register(c *gin.Context) {
	slog.Info("Registration")
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		err := errors.Wrapf(entities.ErrInvalidParam, "decode is failed: %s", err)
		slog.Error("Decode", "err", err)
		s.errProcessing(err, c)
		return
	}

	ctx := c.Request.Context()
	user, err := s.service.CreateUser(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		slog.Error("CreateUser", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "success", "user_id": user.ID})
	return
}

func (s *Server) Login(c *gin.Context) {
	slog.Info("Login")

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		err := errors.Wrapf(entities.ErrInvalidParam, "decode is failed: %s", err)
		slog.Error("Decode", "err", err)
		s.errProcessing(err, c)
		return
	}

	ctx := c.Request.Context()

	user, token, err := s.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		err := errors.Wrap(err, "logging failed")
		slog.Error("Login", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user.ID})
	return
}

func (s *Server) GetTeams(c *gin.Context) {
	slog.Info("GetTeams")
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	teams, err := s.service.GetTeams(ctx, userID)
	if err != nil {
		err := errors.Wrap(err, "getting teams failed")
		slog.Error("GetTeams", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
	return
}

func (s *Server) GetTeamByID(c *gin.Context) {
	slog.Info("GetTeamByID")
	ctx := c.Request.Context()
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing team id failed")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	team, err := s.service.GetTeamByID(ctx, teamID)
	if err != nil {
		err := errors.Wrap(err, "getting team failed")
		slog.Error("GetTeamByID", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
	return
}

func (s *Server) CreateTeam(c *gin.Context) {
	slog.Info("CreateTeam")
	ctx := c.Request.Context()

	var req dto.TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err := errors.Wrapf(entities.ErrInvalidParam, "decode is failed: %s", err)
		slog.Error("Decode", "err", err)
		s.errProcessing(err, c)
		return
	}

	userID := c.GetInt64("user_id")
	team, err := s.service.CreateTeam(ctx, userID, req.Name)
	if err != nil {
		err := errors.Wrap(err, "creating team failed")
		slog.Error("CreateTeam", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": team})
	return
}

func (s *Server) AddMember(c *gin.Context) {
	slog.Info("AddMember")
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")
	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing team id")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	var req dto.TeamInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err := errors.Wrapf(entities.ErrInvalidParam, "decode is failed: %s", err)
		slog.Error("Decode", "err", err)
		s.errProcessing(err, c)
		return
	}

	err = s.service.AddMember(ctx, userID, teamID, req.UserID, req.Role)
	if err != nil {
		err := errors.Wrap(err, "adding member failed")
		slog.Error("AddMember", "err", err)
		s.errProcessing(err, c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "success", "user_id": req.UserID})
	return
}

func (s *Server) CreateTask(c *gin.Context) {
	slog.Info("CreateTask")
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")
	var req dto.TaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		err := errors.Wrapf(entities.ErrInvalidParam, "decode is failed: %s", err)
		slog.Error("Decode", "err", err)
		s.errProcessing(err, c)
		return
	}

	task, err := s.service.CreateTask(ctx, userID, req.AssigneeID, req.TeamID, req.Title, req.Description)
	if err != nil {
		err := errors.Wrap(err, "creating task failed")
		slog.Error("CreateTask", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "success", "task": task})
	return
}

func (s *Server) GetTaskByID(c *gin.Context) {
	slog.Info("GetTaskByID")
	ctx := c.Request.Context()
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing task id failed")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	task, err := s.service.GetTaskByID(ctx, taskID)
	if err != nil {
		err := errors.Wrap(err, "getting task failed")
		slog.Error("GetTaskByID", "err", err)
		s.errProcessing(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
	return
}

func (s *Server) UpdateTaskByID(c *gin.Context) {
	slog.Info("UpdateTaskByID")
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing task id failed")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err := errors.Wrapf(entities.ErrInvalidParam, "decode is failed: %s", err)
		slog.Error("Decode", "err", err)
		s.errProcessing(err, c)
		return
	}
	oldTask, err := s.service.GetTaskByID(ctx, taskID)
	if err != nil {
		err := errors.Wrap(err, "getting task failed")
		slog.Error("GetTaskByID", "err", err)
		s.errProcessing(err, c)
		return
	}

	task := &entities.Task{}

	task.OwnerID = oldTask.OwnerID
	task.ID = oldTask.ID
	task.TeamID = oldTask.TeamID
	task.SetAssigneeID(oldTask.GetAssigneeID())
	task.CreateAt = oldTask.CreateAt
	task.SetID(taskID)
	task.SetTitle(req.Title)
	task.SetDescription(req.Description)
	err = task.SetStatus(req.Status)
	if err != nil {
		err := errors.Wrap(err, "updating task failed")
		slog.Error("SetStatus", "err", err)
		s.errProcessing(err, c)
		return
	}

	err = s.service.UpdateTask(ctx, userID, task)
	if err != nil {
		err := errors.Wrap(err, "updating task failed")
		slog.Error("UpdateTaskByID", "err", err)
		s.errProcessing(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "task": task})
	return
}

func (s *Server) AddComment(c *gin.Context) {
	slog.Info("AddComment")
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")
	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing task id failed")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	var req dto.CommentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		err := errors.Wrapf(entities.ErrInvalidParam, "decode is failed: %s", err)
		slog.Error("Decode", "err", err)
		s.errProcessing(err, c)
		return
	}

	newComm, err := s.service.AddComment(ctx, userID, taskID, req.Comment)
	if err != nil {
		err := errors.Wrap(err, "adding comment failed")
		slog.Error("AddComment", "err", err)
		s.errProcessing(err, c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "success", "task": newComm})
	return
}

func (s *Server) GetHistory(c *gin.Context) {
	slog.Info("GetHistory")
	ctx := c.Request.Context()

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing task id failed")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	history, err := s.service.GetTaskHistory(ctx, taskID, limit, offset)
	if err != nil {
		err := errors.Wrap(err, "getting history failed")
		slog.Error("GetHistory", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (s *Server) GetTeamMembers(c *gin.Context) {
	slog.Info("GetTeamMembers")
	ctx := c.Request.Context()

	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing team id failed")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	members, err := s.service.GetTeamMembers(ctx, teamID)
	if err != nil {
		err := errors.Wrap(err, "getting team members failed")
		slog.Error("GetTeamMembers", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (s *Server) GetTasksByTeam(c *gin.Context) {
	slog.Info("GetTasksByTeam")
	ctx := c.Request.Context()

	slog.Info("Params", "all", c.Params)

	teamID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing team id failed")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	tasks, err := s.service.GetTasksByTeam(ctx, teamID, limit, offset)
	if err != nil {
		err := errors.Wrap(err, "getting tasks by team failed")
		slog.Error("GetTasksByTeam", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (s *Server) GetCommentsByTask(c *gin.Context) {
	slog.Info("GetCommentsByTask")
	ctx := c.Request.Context()

	taskID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		err := errors.Wrap(entities.ErrInvalidParam, "parsing task id failed")
		slog.Error("ParseID", "err", err)
		s.errProcessing(err, c)
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	comments, err := s.service.GetCommentsByTask(ctx, taskID, limit, offset)
	if err != nil {
		err := errors.Wrap(err, "getting comments failed")
		slog.Error("GetCommentsByTask", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func (s *Server) errProcessing(err error, c *gin.Context) {
	errDTO := dto.ErrorDTO{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
	}

	switch {
	case errors.Is(err, entities.ErrInternal):
		errDTO.Status = http.StatusInternalServerError

	case errors.Is(err, entities.ErrInvalidParam):
		errDTO.Status = http.StatusBadRequest

	case errors.Is(err, entities.ErrNotFound):
		errDTO.Status = http.StatusNotFound

	default:
		errDTO.Status = http.StatusInternalServerError
	}

	c.JSON(errDTO.Status, gin.H{"error": errDTO.Message})
}
