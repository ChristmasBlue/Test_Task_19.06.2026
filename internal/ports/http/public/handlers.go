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

// @Summary Регистрация пользователя
// @Description Создаёт нового пользователя в системе
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Данные для регистрации"
// @Success 201 {object} map[string]interface{} "{"message":"success","user_id":1}"
// @Failure 400 {object} dto.ErrorDTO "Неверный запрос"
// @Router /api/v1/register [post]
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
	user, err := s.Service.CreateUser(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		slog.Error("CreateUser", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "success", "user_id": user.ID})
	return
}

// @Summary Вход в систему
// @Description Аутентификация пользователя и получение JWT токена
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Данные для входа"
// @Success 200 {object} map[string]interface{} "{"token":"jwt_token","user":1}"
// @Failure 400 {object} dto.ErrorDTO "Неверные данные"
// @Router /api/v1/login [post]
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

	user, token, err := s.Service.Login(ctx, req.Email, req.Password)
	if err != nil {
		err := errors.Wrap(err, "logging failed")
		slog.Error("Login", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "user": user.ID})
	return
}

// @Summary Список команд пользователя
// @Description Возвращает все команды, в которых состоит пользователь, с его ролью
// @Tags Teams
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} dto.TeamWithRole "Список команд с ролью"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/teams [get]
func (s *Server) GetTeams(c *gin.Context) {
	slog.Info("GetTeams")
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	teams, err := s.Service.GetTeams(ctx, userID)
	if err != nil {
		err := errors.Wrap(err, "getting teams failed")
		slog.Error("GetTeams", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
	return
}

// @Summary Получение команды по ID
// @Description Возвращает информацию о команде
// @Tags Teams
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID команды"
// @Success 200 {object} map[string]interface{} "{"team":{...}}"
// @Failure 404 {object} dto.ErrorDTO "Команда не найдена"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/teams/{id} [get]
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

	team, err := s.Service.GetTeamByID(ctx, teamID)
	if err != nil {
		err := errors.Wrap(err, "getting team failed")
		slog.Error("GetTeamByID", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
	return
}

// @Summary Создание команды
// @Description Создаёт новую команду и назначает создателя владельцем
// @Tags Teams
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.TeamRequest true "Название команды"
// @Success 201 {object} map[string]interface{} "{"team":{...}}"
// @Failure 400 {object} dto.ErrorDTO "Неверный запрос"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/teams [post]
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
	team, err := s.Service.CreateTeam(ctx, userID, req.Name)
	if err != nil {
		err := errors.Wrap(err, "creating team failed")
		slog.Error("CreateTeam", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"team": team})
	return
}

// @Summary Пригласить пользователя в команду
// @Description Добавляет пользователя в команду (только owner/admin)
// @Tags Teams
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID команды"
// @Param request body dto.TeamInviteRequest true "ID пользователя и роль"
// @Success 201 {object} map[string]interface{} "{"message":"success","user_id":2}"
// @Failure 400 {object} dto.ErrorDTO "Неверный запрос"
// @Failure 403 {object} dto.ErrorDTO "Недостаточно прав"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/teams/{id}/invite [post]
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

	err = s.Service.AddMember(ctx, userID, teamID, req.UserID, req.Role)
	if err != nil {
		err := errors.Wrap(err, "adding member failed")
		slog.Error("AddMember", "err", err)
		s.errProcessing(err, c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "success", "user_id": req.UserID})
	return
}

// @Summary Создание задачи
// @Description Создаёт новую задачу в указанной команде
// @Tags Tasks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.TaskRequest true "Данные задачи"
// @Success 201 {object} map[string]interface{} "{"message":"success","task":{...}}"
// @Failure 400 {object} dto.ErrorDTO "Неверный запрос"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/tasks [post]
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

	task, err := s.Service.CreateTask(ctx, userID, req.AssigneeID, req.TeamID, req.Title, req.Description)
	if err != nil {
		err := errors.Wrap(err, "creating task failed")
		slog.Error("CreateTask", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "success", "task": task})
	return
}

// @Summary Получение задачи по ID
// @Description Возвращает задачу по её идентификатору
// @Tags Tasks
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID задачи"
// @Success 200 {object} map[string]interface{} "{"task":{...}}"
// @Failure 404 {object} dto.ErrorDTO "Задача не найдена"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/tasks/{id} [get]
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

	task, err := s.Service.GetTaskByID(ctx, taskID)
	if err != nil {
		err := errors.Wrap(err, "getting task failed")
		slog.Error("GetTaskByID", "err", err)
		s.errProcessing(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"task": task})
	return
}

// @Summary Обновление задачи
// @Description Обновляет задачу (можно менять title, description, status)
// @Tags Tasks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID задачи"
// @Param request body dto.UpdateTaskRequest true "Данные для обновления"
// @Success 200 {object} map[string]interface{} "{"message":"success","task":{...}}"
// @Failure 400 {object} dto.ErrorDTO "Неверный запрос"
// @Failure 404 {object} dto.ErrorDTO "Задача не найдена"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/tasks/{id} [put]
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
	oldTask, err := s.Service.GetTaskByID(ctx, taskID)
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

	err = s.Service.UpdateTask(ctx, userID, task)
	if err != nil {
		err := errors.Wrap(err, "updating task failed")
		slog.Error("UpdateTaskByID", "err", err)
		s.errProcessing(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success", "task": task})
	return
}

// @Summary Добавление комментария к задаче
// @Description Добавляет комментарий к задаче (только член команды)
// @Tags Comments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID задачи"
// @Param request body dto.CommentRequest true "Текст комментария"
// @Success 201 {object} map[string]interface{} "{"message":"success","task":{...}}"
// @Failure 400 {object} dto.ErrorDTO "Неверный запрос"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/tasks/{id}/comments [post]
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

	newComm, err := s.Service.AddComment(ctx, userID, taskID, req.Comment)
	if err != nil {
		err := errors.Wrap(err, "adding comment failed")
		slog.Error("AddComment", "err", err)
		s.errProcessing(err, c)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "success", "task": newComm})
	return
}

// @Summary История изменений задачи
// @Description Возвращает историю изменений задачи с пагинацией
// @Tags History
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID задачи"
// @Param limit query int false "Количество записей (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {array} entities.TaskHistory "Список изменений"
// @Failure 404 {object} dto.ErrorDTO "Задача не найдена"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/tasks/{id}/history [get]
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

	history, err := s.Service.GetTaskHistory(ctx, taskID, limit, offset)
	if err != nil {
		err := errors.Wrap(err, "getting history failed")
		slog.Error("GetHistory", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

// @Summary Список участников команды
// @Description Возвращает всех участников команды с их ролями
// @Tags Teams
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID команды"
// @Success 200 {array} entities.TeamMember "Список участников"
// @Failure 404 {object} dto.ErrorDTO "Команда не найдена"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/teams/{id}/members [get]
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

	members, err := s.Service.GetTeamMembers(ctx, teamID)
	if err != nil {
		err := errors.Wrap(err, "getting team members failed")
		slog.Error("GetTeamMembers", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// @Summary Задачи команды
// @Description Возвращает все задачи указанной команды с пагинацией
// @Tags Tasks
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID команды"
// @Param limit query int false "Количество записей (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {array} entities.Task "Список задач"
// @Failure 404 {object} dto.ErrorDTO "Команда не найдена"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/teams/{id}/tasks [get]
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

	tasks, err := s.Service.GetTasksByTeam(ctx, teamID, limit, offset)
	if err != nil {
		err := errors.Wrap(err, "getting tasks by team failed")
		slog.Error("GetTasksByTeam", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// @Summary Получение комментариев к задаче
// @Description Возвращает все комментарии к задаче с пагинацией
// @Tags Comments
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "ID задачи"
// @Param limit query int false "Количество записей (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {array} entities.TaskComment "Список комментариев"
// @Failure 404 {object} dto.ErrorDTO "Задача не найдена"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/tasks/{id}/comments [get]
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

	comments, err := s.Service.GetCommentsByTask(ctx, taskID, limit, offset)
	if err != nil {
		err := errors.Wrap(err, "getting comments failed")
		slog.Error("GetCommentsByTask", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

// @Summary Список задач с фильтрацией
// @Description Возвращает задачи с фильтрацией по команде, статусу, исполнителю и пагинацией
// @Tags Tasks
// @Produce json
// @Security ApiKeyAuth
// @Param team_id query int false "ID команды"
// @Param status query string false "Статус задачи (todo, in_progress, done)"
// @Param assignee_id query int false "ID исполнителя"
// @Param limit query int false "Количество записей (по умолчанию 20)"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Success 200 {array} entities.Task "Список задач"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/tasks [get]
func (s *Server) GetTasksByFilter(c *gin.Context) {
	slog.Info("GetTaskByFilter")
	ctx := c.Request.Context()
	userID := c.GetInt64("user_id")

	var filter dto.TaskFilter

	if teamIDStr := c.Query("team_id"); teamIDStr != "" {
		teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
		if err == nil {
			filter.TeamIDs = []int64{teamID}
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		filter.Status = &statusStr
	}

	if assigneeStr := c.Query("assignee_id"); assigneeStr != "" {
		assignee, err := strconv.ParseInt(assigneeStr, 10, 64)
		if err == nil {
			filter.AssigneeID = &assignee
		}
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	filter.Limit = limit

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	filter.Offset = offset

	tasks, err := s.Service.GetTasksByFilter(ctx, userID, filter)
	if err != nil {
		slog.Error("GetTasks", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// @Summary Статистика по командам
// @Description Возвращает статистику по каждой команде: количество участников и выполненных задач за 7 дней
// @Tags Reports
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} dto.TeamStats "Статистика по командам"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/reports/team-stats [get]
func (s *Server) GetTeamStats(c *gin.Context) {
	slog.Info("GetTeamStats")
	ctx := c.Request.Context()

	stats, err := s.Service.GetTeamStats(ctx)
	if err != nil {
		slog.Error("GetTeamStats", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// @Summary Топ создателей задач
// @Description Возвращает топ-3 пользователей по количеству созданных задач в каждой команде за месяц
// @Tags Reports
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} dto.TopCreator "Топ создателей задач"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/reports/top-creators [get]
func (s *Server) GetTopCreators(c *gin.Context) {
	slog.Info("GetTopCreators")
	ctx := c.Request.Context()

	topCreators, err := s.Service.GetTopCreators(ctx)
	if err != nil {
		slog.Error("GetTopCreators", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"topCreators": topCreators})
}

// @Summary Задачи с невалидным исполнителем
// @Description Возвращает задачи, где исполнитель не является членом команды
// @Tags Reports
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {array} dto.InvalidAssigneeTask "Задачи с невалидным исполнителем"
// @Failure 401 {object} dto.ErrorDTO "Неавторизован"
// @Router /api/v1/reports/invalid-assignee [get]
func (s *Server) GetInvalidAssignee(c *gin.Context) {
	slog.Info("GetInvalidAssigneeTasks")
	ctx := c.Request.Context()

	invalidAssigneeTasks, err := s.Service.GetInvalidAssigneeTasks(ctx)
	if err != nil {
		slog.Error("GetInvalidAssigneeTasks", "err", err)
		s.errProcessing(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"invalidAssigneeTasks": invalidAssigneeTasks})
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
