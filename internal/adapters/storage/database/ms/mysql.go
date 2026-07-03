package ms

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"test_task/internal/cases"
	"test_task/internal/entities"
	"test_task/pkg/dto"
	"test_task/tools/config"
	"test_task/tools/executor"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/sync/singleflight"

	"github.com/pkg/errors"
)

type Storage struct {
	db   *sql.DB
	exec executor.Executor
	sf   *singleflight.Group
}

//var _ cases.Repository = (Storage)(nil)

func NewStorage(cfg config.Config) (*Storage, error) {
	db, err := sql.Open("mysql", cfg.StorageConnStr())
	if err != nil {
		slog.Error("Open mysql", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, "open mysql connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		slog.Error("Open mysql", "err", err)
		db.Close()
		return nil, errors.Wrap(entities.ErrInternal, "open mysql connection")
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns())
	db.SetMaxIdleConns(cfg.MaxIdleConns())
	db.SetConnMaxLifetime(cfg.LifeConns())
	db.SetConnMaxIdleTime(cfg.LifeIdleConns())

	return &Storage{
		db: db,
		sf: &singleflight.Group{},
	}, nil
}

func (s *Storage) Start() error {
	return nil
}

func (s *Storage) Stop(ctx context.Context) error {
	<-ctx.Done()
	if s.db == nil {
		return nil
	}

	slog.Info("Closing database connection")
	return s.db.Close()
}

func (s *Storage) Ping(ctx context.Context) error {
	if s.db == nil {
		return errors.Wrap(entities.ErrInternal, "database connection is nil")
	}

	slog.Info("Ping database")
	return s.db.PingContext(ctx)
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {

	key := fmt.Sprintf("db:user:email:%s", email)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetUserByEmail")
		query := `SELECT id, name, hash_pass, create_at 
				FROM users
				WHERE email = ?`

		e := s.execution()

		var (
			userID                 int64
			userName, userHashPass string
			userCreateAt           time.Time
		)

		err := e.QueryRowContext(ctx, query, email).Scan(&userID, &userName, &userHashPass, &userCreateAt)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.Wrap(entities.ErrNotFound, "get user by email")
			}
			slog.Error("QueryRowContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetUserByEmail")
		}

		user, err := entities.NewUser(userID, userName, email, []byte(userHashPass), userCreateAt)
		if err != nil {
			slog.Error("NewUSer", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetUserByEmail")
		}

		return user, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.User), nil
}

func (s *Storage) GetUserByID(ctx context.Context, id int64) (*entities.User, error) {
	key := fmt.Sprintf("db:user:id:%d", id)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetUserByID")

		query := `SELECT name, email, hash_pass, create_at 
				FROM users
				WHERE id = ?`

		e := s.execution()

		var (
			userEmail, userName, userHashPass string
			userCreateAt                      time.Time
		)

		err := e.QueryRowContext(ctx, query, id).Scan(&userName, &userEmail, &userHashPass, &userCreateAt)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.Wrap(entities.ErrNotFound, "get user by id")
			}
			slog.Error("QueryRowContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetUserByID")
		}

		user, err := entities.NewUser(id, userName, userEmail, []byte(userHashPass), userCreateAt)
		if err != nil {
			slog.Error("NewUSer", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetUserByID")
		}

		return user, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.User), nil
}

func (s *Storage) CreateUser(ctx context.Context, name, email, hashPass string) (*entities.User, error) {
	slog.Info("CreateUser")
	userCreateAt := time.Now()
	query := `INSERT INTO users (name, email, hash_pass, create_at) 
				VALUES (?, ?, ?, ?)`

	e := s.execution()

	result, err := e.ExecContext(ctx, query, name, email, hashPass, userCreateAt)
	if err != nil {
		slog.Error("ExecContext", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, "failed to create user")
	}
	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("LastInsertId", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, "failed to create user")
	}

	user, err := entities.NewUser(id, name, email, []byte(hashPass), userCreateAt)
	if err != nil {
		slog.Error("NewUser", "err", err)
		return nil, errors.Wrap(err, "create user")
	}

	return user, nil
}

func (s *Storage) AddMember(ctx context.Context, userID, teamID int64, role string) error {
	slog.Info("AddMember")
	query := `INSERT INTO team_members (user_id, team_id, role) VALUES (?, ?, ?)`

	e := s.execution()

	_, err := e.ExecContext(ctx, query, userID, teamID, role)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return entities.ErrConflict
		}
		slog.Error("ExecContext", "err", err)
		return errors.Wrap(entities.ErrInternal, err.Error())
	}

	return nil
}

func (s *Storage) CreateTeam(ctx context.Context, name string, ownerID int64) (*entities.Team, error) {
	slog.Info("CreateTeam")
	teamCreateAt := time.Now()

	query := `INSERT INTO teams (name, owner_id, create_at) 
				VALUES (?, ?, ?)`

	e := s.execution()

	result, err := e.ExecContext(ctx, query, name, ownerID, teamCreateAt)
	if err != nil {
		slog.Error("ExecContext", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("LastInsertId", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	team, err := entities.NewTeam(id, name, ownerID, teamCreateAt)
	if err != nil {
		slog.Error("NewTeam", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	return team, nil
}

func (s *Storage) GetTeamsByIDs(ctx context.Context, ids []int64) ([]*entities.Team, error) {

	sorted := make([]int64, len(ids))
	copy(sorted, ids)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	key := "db:teams:ids:"
	for i, id := range sorted {
		if i > 0 {
			key += ","
		}
		key += fmt.Sprintf("%d", id)
	}

	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetTeamsByIDs")

		e := s.execution()

		if len(ids) == 0 {
			return []*entities.Team{}, nil
		}

		placeholders := strings.Repeat("?,", len(ids))
		placeholders = placeholders[:len(placeholders)-1]

		query := `SELECT id, name, owner_id, create_at 
				FROM teams
				WHERE id IN (` + placeholders + `)`

		args := make([]interface{}, len(ids))
		for i, id := range ids {
			args[i] = id
		}

		rows, err := e.QueryContext(ctx, query, args...)
		if err != nil {
			slog.Error("QueryContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTeamsByIDs")
		}
		defer rows.Close()

		var teams []*entities.Team

		var (
			teamID     int64
			teamName   string
			teamOwner  int64
			teamCreate time.Time
		)

		for rows.Next() {
			if err := rows.Scan(&teamID, &teamName, &teamOwner, &teamCreate); err != nil {
				slog.Error("Scan", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetTeamsByIDs")
			}

			team, err := entities.NewTeam(teamID, teamName, teamOwner, teamCreate)
			if err != nil {
				slog.Error("NewTeam", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetTeamsByIDs")
			}

			teams = append(teams, team)
		}

		if err := rows.Err(); err != nil {
			slog.Error("Rows", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTeamsByIDs")
		}

		return teams, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*entities.Team), nil
}

func (s *Storage) GetTeamByID(ctx context.Context, id int64) (*entities.Team, error) {
	key := fmt.Sprintf("db:team:id:%d", id)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetTeamByID")

		query := `SELECT name, owner_id, create_at 
				FROM teams 
				WHERE id = ?`

		e := s.execution()

		var (
			teamName     string
			teamOwnerID  int64
			teamCreateAt time.Time
		)

		err := e.QueryRowContext(ctx, query, id).Scan(&teamName, &teamOwnerID, &teamCreateAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.Wrap(entities.ErrNotFound, "team not found")
			}
			slog.Error("QueryRowContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTeamByID")
		}

		team, err := entities.NewTeam(id, teamName, teamOwnerID, teamCreateAt)
		if err != nil {
			slog.Error("NewTeam", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTeamByID")
		}

		return team, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.Team), nil
}

func (s *Storage) GetUserTeams(ctx context.Context, userID int64) ([]int64, error) {
	key := fmt.Sprintf("db:user_teams:id:%d", userID)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetUserTeams")

		query := `SELECT team_id 
				FROM team_members 
				WHERE user_id = ?`

		e := s.execution()

		rows, err := e.QueryContext(ctx, query, userID)
		if err != nil {
			slog.Error("QueryContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetUserTeams")
		}
		defer rows.Close()

		var teamIDs []int64
		for rows.Next() {
			var teamID int64
			if err := rows.Scan(&teamID); err != nil {
				slog.Error("Scan", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetUserTeams")
			}
			teamIDs = append(teamIDs, teamID)
		}

		if err := rows.Err(); err != nil {
			slog.Error("Rows", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetUserTeams")
		}

		return teamIDs, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]int64), nil
}

func (s *Storage) GetTeamMembers(ctx context.Context, teamID int64) ([]*entities.TeamMember, error) {
	key := fmt.Sprintf("db:team_members:id:%d", teamID)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetTeamMembers")

		query := `SELECT user_id, role FROM team_members WHERE team_id = ?`

		e := s.execution()

		rows, err := e.QueryContext(ctx, query, teamID)
		if err != nil {
			slog.Error("QueryContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTeamMembers")
		}
		defer rows.Close()

		var members []*entities.TeamMember

		var (
			userID int64
			role   string
		)

		for rows.Next() {
			if err := rows.Scan(&userID, &role); err != nil {
				slog.Error("Scan", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetTeamMembers")
			}

			member, err := entities.NewTeamMember(userID, teamID, role)
			if err != nil {
				slog.Error("NewTeamMember", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetTeamMembers")
			}

			members = append(members, member)
		}

		if err := rows.Err(); err != nil {
			slog.Error("Rows", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTeamMembers")
		}

		return members, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]*entities.TeamMember), nil
}

func (s *Storage) IsAdminOrOwner(ctx context.Context, userID, teamID int64) (bool, error) {
	key := fmt.Sprintf("db:is_admin:id:%d,%d", userID, teamID)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("IsAdminOrOwner")

		query := `SELECT role FROM team_members WHERE user_id = ? AND team_id = ?`

		e := s.execution()

		var role string

		err := e.QueryRowContext(ctx, query, userID, teamID).Scan(&role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, errors.Wrap(entities.ErrNotFound, "user not found")
			}
			slog.Error("QueryRowContext", "err", err)
			return false, errors.Wrap(entities.ErrInternal, "IsAdminOrOwner")
		}

		return role == "admin" || role == "owner", nil
	})
	if err != nil {
		return false, err
	}

	return result.(bool), nil
}

func (s *Storage) IsMember(ctx context.Context, userID, teamID int64) (bool, error) {
	key := fmt.Sprintf("db:is_member:id:%d,%d", userID, teamID)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("IsMember", "user_id", userID, "team_id", teamID)

		e := s.execution()

		query := `SELECT EXISTS(SELECT 1 FROM team_members WHERE user_id = ? AND team_id = ?)`

		var exists bool
		err := e.QueryRowContext(ctx, query, userID, teamID).Scan(&exists)
		if err != nil {
			slog.Error("QueryRowContext", "err", err)
			return false, errors.Wrap(entities.ErrInternal, "failed to check membership")
		}

		return exists, nil
	})
	if err != nil {
		return false, err
	}

	return result.(bool), nil
}

func (s *Storage) GetMemberRole(ctx context.Context, userID, teamID int64) (string, error) {
	key := fmt.Sprintf("db:member_role:id:%d,%d", userID, teamID)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetMemberRole", "user_id", userID, "team_id", teamID)

		e := s.execution()

		query := `SELECT role FROM team_members WHERE user_id = ? AND team_id = ?`

		var role string
		err := e.QueryRowContext(ctx, query, userID, teamID).Scan(&role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", errors.Wrap(entities.ErrNotFound, "user is not a member of this team")
			}
			slog.Error("QueryRowContext", "err", err)
			return "", errors.Wrap(entities.ErrInternal, "failed to get member role")
		}

		return role, nil
	})
	if err != nil {
		return "", err
	}

	return result.(string), nil
}

func (s *Storage) CreateTask(ctx context.Context, userID, assigneeID, teamID int64, title, description, status string) (*entities.Task, error) {
	slog.Info("CreateTask")

	query := `INSERT INTO tasks (team_id, owner_id, title, description, status, assignee_id, create_at) 
				VALUES (?, ?, ?, ?, ?, ?, ?)`

	e := s.execution()

	now := time.Now()

	if status == "" {
		status = entities.TaskStatusTODO
	}

	result, err := e.ExecContext(ctx, query,
		teamID,
		userID,
		title,
		description,
		status,
		assigneeID,
		now,
	)
	if err != nil {
		slog.Error("ExecContext", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, "CreateTask")
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("LastInsertId", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, "CreateTask")
	}

	task, err := entities.NewTask(id, userID, teamID, assigneeID, title, description, status, now)
	if err != nil {
		slog.Error("NewTask", "err", err)
		return nil, errors.Wrap(err, "CreateTask")
	}

	return task, nil
}

func (s *Storage) GetTaskByID(ctx context.Context, id int64) (*entities.Task, error) {
	key := fmt.Sprintf("db:task:id:%d", id)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetTaskByID")

		query := `SELECT id, team_id, owner_id, title, description, status, assignee_id, create_at, complete_at
				FROM tasks
				WHERE id = ?`

		e := s.execution()

		var (
			taskID, teamID, ownerID, assigneeID int64
			title, description, status          string
			createAt                            time.Time
			completeAt                          sql.NullTime
		)

		err := e.QueryRowContext(ctx, query, id).Scan(
			&taskID,
			&teamID,
			&ownerID,
			&title,
			&description,
			&status,
			&assigneeID,
			&createAt,
			&completeAt,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.Wrap(entities.ErrNotFound, "task not found")
			}
			slog.Error("QueryRowContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTaskByID")
		}

		task, err := entities.NewTask(
			taskID,
			ownerID,
			teamID,
			assigneeID,
			title,
			description,
			status,
			createAt,
		)
		if err != nil {
			slog.Error("NewTask", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTaskByID")
		}

		var completeAtPtr *time.Time
		if completeAt.Valid {
			completeAtPtr = &completeAt.Time
		}

		task.SetCompeteAt(completeAtPtr)

		return task, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*entities.Task), nil
}

func (s *Storage) GetTasksByTeam(ctx context.Context, teamID int64, limit, offset int) ([]*entities.Task, error) {
	key := fmt.Sprintf("db:task_by_team:id:%d:limit:%d:offset:%d", teamID, limit, offset)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetTasksByTeam")

		query := `SELECT id, team_id, owner_id, title, description, status, assignee_id, create_at, complete_at
				FROM tasks
				WHERE team_id = ?
				ORDER BY id DESC`

		e := s.execution()

		args := []interface{}{}

		args = append(args, teamID)

		if limit != 0 {
			query += `LIMIT ? OFFSET ?`
			args = append(args, limit, offset)
		}

		rows, err := e.QueryContext(ctx, query, args...)
		if err != nil {
			slog.Error("QueryContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTasksByTeam")
		}
		defer rows.Close()

		var (
			taskID, taskTeamID, ownerID, assigneeID int64
			title, description, status              string
			createAt                                time.Time
			completeAt                              sql.NullTime
		)

		var tasks []*entities.Task
		for rows.Next() {

			if err := rows.Scan(
				&taskID,
				&taskTeamID,
				&ownerID,
				&title,
				&description,
				&status,
				&assigneeID,
				&createAt,
				&completeAt,
			); err != nil {
				slog.Error("Scan", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetTasksByTeam")
			}

			task, err := entities.NewTask(
				taskID,
				ownerID,
				taskTeamID,
				assigneeID,
				title,
				description,
				status,
				createAt,
			)
			if err != nil {
				slog.Error("NewTask", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetTasksByTeam")
			}

			var completeAtPtr *time.Time
			if completeAt.Valid {
				completeAtPtr = &completeAt.Time
			}

			task.SetCompeteAt(completeAtPtr)

			tasks = append(tasks, task)
		}

		if err := rows.Err(); err != nil {
			slog.Error("Rows", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTasksByTeam")
		}

		return tasks, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]*entities.Task), nil
}

func (s *Storage) GetTasksByFilter(ctx context.Context, filter dto.TaskFilter) ([]*entities.Task, error) {
	slog.Info("GetTasksByFilter")

	e := s.execution()

	query := `SELECT id, team_id, owner_id, title, description, status, assignee_id, create_at, complete_at
              FROM tasks 
              WHERE 1=1`

	args := []interface{}{}

	if len(filter.TeamIDs) > 0 {
		placeholders := strings.Repeat("?,", len(filter.TeamIDs))
		placeholders = placeholders[:len(placeholders)-1]
		query += ` AND team_id IN (` + placeholders + `)`
		for _, id := range filter.TeamIDs {
			args = append(args, id)
		}
	}

	if filter.Status != nil && *filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, *filter.Status)
	}

	if filter.AssigneeID != nil {
		query += ` AND assignee_id = ?`
		args = append(args, *filter.AssigneeID)
	}

	query += ` ORDER BY created_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := e.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("QueryContext", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}
	defer rows.Close()

	var tasks []*entities.Task
	for rows.Next() {
		var (
			taskID, teamID, ownerID, assigneeID int64
			title, description, status          string
			createAt                            time.Time
			completeAt                          sql.NullTime
		)

		if err := rows.Scan(
			&taskID,
			&teamID,
			&ownerID,
			&title,
			&description,
			&status,
			&assigneeID,
			&createAt,
			&completeAt,
		); err != nil {
			slog.Error("Scan", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}

		task, err := entities.NewTask(
			taskID,
			ownerID,
			teamID,
			assigneeID,
			title,
			description,
			status,
			createAt,
		)

		if err != nil {
			slog.Error("NewTask", "err", err)
			return nil, errors.Wrap(err, "failed to create task entity")
		}

		var completeAtPtr *time.Time
		if completeAt.Valid {
			completeAtPtr = &completeAt.Time
		}

		task.SetCompeteAt(completeAtPtr)

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Rows", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	return tasks, nil
}

func (s *Storage) UpdateTask(ctx context.Context, task *entities.Task) error {
	slog.Info("UpdateTask")

	query := `UPDATE tasks 
				SET title = ?, description = ?, status = ?, assignee_id = ?, complete_at = ?
				WHERE id = ?`

	e := s.execution()

	result, err := e.ExecContext(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.AssigneeID,
		task.CompleteAt,
		task.ID,
	)
	if err != nil {
		slog.Error("ExecContext", "err", err)
		return errors.Wrap(entities.ErrInternal, "UpdateTask")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("RowsAffected", "err", err)
		return errors.Wrap(entities.ErrInternal, "UpdateTask")
	}

	if rowsAffected == 0 {
		return errors.Wrap(entities.ErrNotFound, "task not found")
	}

	return nil
}

func (s *Storage) AddHistoryRecord(ctx context.Context, record *entities.TaskHistory) error {
	slog.Info("AddHistoryRecord")

	query := `INSERT INTO task_history (task_id, change_by, field, old_value, new_value, change_at) 
				VALUES (?, ?, ?, ?, ?, ?)`

	e := s.execution()

	now := time.Now()
	record.ChangeAt = now

	_, err := e.ExecContext(ctx, query,
		record.TaskID,
		record.ChangeBy,
		record.Field,
		record.OldValue,
		record.NewValue,
		now,
	)

	if err != nil {
		slog.Error("ExecContext", "err", err)
		return errors.Wrap(entities.ErrInternal, "AddHistoryRecord")
	}

	return nil
}

func (s *Storage) GetTaskHistory(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error) {
	key := fmt.Sprintf("db:history:id:%d:limit:%d:offset:%d", taskID, limit, offset)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetTaskHistory")

		query := `SELECT change_by, field, old_value, new_value, change_at 
				FROM task_history
				WHERE task_id = ?
				ORDER BY change_at DESC
				LIMIT ? OFFSET ?`

		e := s.execution()

		rows, err := e.QueryContext(ctx, query, taskID, limit, offset)
		if err != nil {
			slog.Error("QueryContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTaskHistory")
		}
		defer rows.Close()

		var history []*entities.TaskHistory

		var (
			changedBy int64
			field     string
			oldValue  *string
			newValue  *string
			changedAt time.Time
		)

		for rows.Next() {

			if err := rows.Scan(&changedBy, &field, &oldValue, &newValue, &changedAt); err != nil {
				slog.Error("Scan", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, err.Error())
			}

			record, err := entities.NewTaskHistory(taskID, changedBy, field, oldValue, newValue, changedAt)
			if err != nil {
				slog.Error("NewTaskHistory", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetTaskHistory")
			}

			history = append(history, record)
		}

		if err := rows.Err(); err != nil {
			slog.Error("Rows", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetTaskHistory")
		}

		return history, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]*entities.TaskHistory), nil
}

func (s *Storage) AddComment(ctx context.Context, comment *entities.TaskComment) error {
	slog.Info("AddComment")

	query := `INSERT INTO task_comments (task_id, user_id, content, create_at) 
				VALUES (?, ?, ?, ?)`

	e := s.execution()

	now := time.Now()
	comment.CreateAt = now

	_, err := e.ExecContext(ctx, query,
		comment.TaskID,
		comment.UserID,
		comment.Content,
		now,
	)
	if err != nil {
		slog.Error("ExecContext", "err", err)
		return errors.Wrap(entities.ErrInternal, "AddComment")
	}

	return nil
}

func (s *Storage) GetCommentsByTask(ctx context.Context, taskID int64, limit, offset int) ([]*entities.TaskComment, error) {
	key := fmt.Sprintf("db:comm_by_task:id:%d:limit:%d:offset:%d", taskID, limit, offset)
	result, err, _ := s.sf.Do(key, func() (interface{}, error) {
		slog.Info("GetCommentsByTask")

		query := `SELECT user_id, content, create_at 
				FROM task_comments
				WHERE task_id = ?
				ORDER BY create_at DESC
				LIMIT ? OFFSET ?`

		e := s.execution()

		rows, err := e.QueryContext(ctx, query, taskID, limit, offset)
		if err != nil {
			slog.Error("QueryContext", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetCommentsByTask")
		}
		defer rows.Close()

		var comments []*entities.TaskComment

		var (
			userID   int64
			content  string
			createAt time.Time
		)

		for rows.Next() {

			if err := rows.Scan(&userID, &content, &createAt); err != nil {
				slog.Error("Scan", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetCommentsByTask")
			}

			comment, err := entities.NewTaskComment(taskID, userID, content, createAt)
			if err != nil {
				slog.Error("NewTaskComment", "err", err)
				return nil, errors.Wrap(entities.ErrInternal, "GetCommentsByTask")
			}

			comments = append(comments, comment)
		}

		if err := rows.Err(); err != nil {
			slog.Error("Rows", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, "GetCommentsByTask")
		}

		return comments, nil
	})
	if err != nil {
		return nil, err
	}

	return result.([]*entities.TaskComment), nil
}

func (s *Storage) getIDs(ctx context.Context, query string, args ...interface{}) ([]int64, error) {
	result := make([]int64, 0)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("QueryContext", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}
	defer rows.Close()
	var id int64
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			slog.Error("Scan", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}
		result = append(result, id)
	}
	if err := rows.Err(); err != nil {
		slog.Error("Rows", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}
	return result, nil
}

func (s *Storage) InTx(ctx context.Context, fn func(ctx context.Context, stor cases.Repository) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		slog.Error("InTx", "err", err)
		return entities.ErrInternal
	}

	defer tx.Rollback()

	newStor := &Storage{
		db:   s.db,
		exec: tx,
	}

	err = fn(ctx, newStor)
	if err != nil {
		slog.Error("InTx", "err", err)

		return err
	}

	if err := tx.Commit(); err != nil {
		slog.Error("Commit", "err", err)
		return errors.Wrap(entities.ErrInternal, "failed to commit transaction")
	}

	return nil
}

func (s *Storage) GetTeamStats(ctx context.Context) ([]dto.TeamStats, error) {
	slog.Info("GetTeamStats")
	e := s.execution()

	query := `SELECT 
    				t.id AS team_id,
    				t.name AS team_name,
    				COUNT(DISTINCT tm.user_id) AS member_count,
    				COUNT(DISTINCT tk.id) AS done_tasks_last_7_days
				FROM teams t
				LEFT JOIN team_members tm ON t.id = tm.team_id
				LEFT JOIN tasks tk ON t.id = tk.team_id 
    				AND tk.status = 'done' 
    				AND tk.complete_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
				GROUP BY t.id, t.name`

	rows, err := e.QueryContext(ctx, query)
	if err != nil {
		slog.Error("QueryContext", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}
	defer rows.Close()

	var stats []dto.TeamStats
	for rows.Next() {
		var stat dto.TeamStats
		if err := rows.Scan(&stat.TeamID, &stat.TeamName, &stat.MemberCount, &stat.DoneTasksLast7); err != nil {
			slog.Error("Scan", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

func (s *Storage) GetTopCreators(ctx context.Context) ([]dto.TopCreator, error) {
	slog.Info("GetTopCreators")
	e := s.execution()

	query := `
        SELECT team_id, team_name, user_id, user_name, task_count, rank
		FROM (SELECT t.id AS team_id, t.name AS team_name, u.id AS user_id, u.name AS user_name, COUNT(tk.id) AS task_count, RANK() OVER (PARTITION BY t.id ORDER BY COUNT(tk.id) DESC) AS rank
    			FROM teams t
    			JOIN tasks tk ON t.id = tk.team_id
    			JOIN users u ON tk.owner_id = u.id
    			WHERE tk.create_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH)
    			GROUP BY t.id, t.name, u.id, u.name
			) ranked
		WHERE rank <= 3
		ORDER BY team_id, rank`

	rows, err := e.QueryContext(ctx, query)
	if err != nil {
		slog.Error("QueryContext", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}
	defer rows.Close()

	var creators []dto.TopCreator
	for rows.Next() {
		var c dto.TopCreator
		if err := rows.Scan(&c.TeamID, &c.TeamName, &c.UserID, &c.UserName, &c.TaskCount, &c.Rank); err != nil {
			slog.Error("Scan", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}
		creators = append(creators, c)
	}

	return creators, rows.Err()
}

func (s *Storage) GetInvalidAssigneeTasks(ctx context.Context) ([]dto.InvalidAssigneeTask, error) {
	slog.Info("GetInvalidAssigneeTasks")
	e := s.execution()

	query := `
        SELECT 
            tk.id AS task_id,
            tk.title AS task_title,
            t.id AS team_id,
            t.name AS team_name,
            u.id AS assignee_id,
            u.name AS assignee_name
        FROM tasks tk
        JOIN teams t ON tk.team_id = t.id
        JOIN users u ON tk.assignee_id = u.id
        LEFT JOIN team_members tm ON t.id = tm.team_id AND u.id = tm.user_id
        WHERE tk.assignee_id IS NOT NULL
          AND tm.user_id IS NULL`

	rows, err := e.QueryContext(ctx, query)
	if err != nil {
		slog.Error("QueryContext", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}
	defer rows.Close()

	var tasks []dto.InvalidAssigneeTask
	for rows.Next() {
		var task dto.InvalidAssigneeTask
		if err := rows.Scan(&task.TaskID, &task.TaskTitle, &task.TeamID, &task.TeamName, &task.AssigneeID, &task.AssigneeName); err != nil {
			slog.Error("Scan", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *Storage) execution() executor.Executor {
	if s.exec != nil {
		return s.exec
	}
	return s.db
}
