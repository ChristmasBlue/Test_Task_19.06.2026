package pg

import (
	"context"
	"log/slog"
	"sync"
	"test_task/internal/cases"
	"test_task/internal/entities"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

var _ cases.Storage = (*Storage)(nil)

type Storage struct {
	db       *pgxpool.Pool
	cancelFn context.CancelFunc
	once     sync.Once
}

func NewStorage(connString string) (*Storage, error) {
	if connString == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "connString is empty")
	}

	ctx, cancel := context.WithCancel(context.Background())

	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		cancel()
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	return &Storage{
		db:       db,
		cancelFn: cancel,
	}, nil
}

func (s *Storage) Close() {
	s.once.Do(func() {
		s.cancelFn()
		s.db.Close()
	})
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	slog.Info("GetUserByEmail")
	query := `SELECT id, name,  email, hash_pass, create_at FROM users WHERE email = $1 LIMIT 1`

	query2 := `SELECT team_id FROM teams_users WHERE user_id = $1`

	user := &entities.User{}

	err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.HashPassword, &user.CreateAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(entities.ErrNotFound, "user not found")
		}

		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	teamsID, err := s.getIDs(ctx, query2, user.ID)
	if err != nil {
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(err, "get teams ID")
	}

	user.TeamsID = teamsID

	return user, nil
}

func (s *Storage) GetUserByID(ctx context.Context, id int64) (*entities.User, error) {
	slog.Info("GetUserByID")
	query := `SELECT id, name,  email, hash_pass, create_at FROM users WHERE id = $1 LIMIT 1`

	query2 := `SELECT team_id FROM teams_users WHERE user_id = $1`

	user := &entities.User{}

	err := s.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.HashPassword, &user.CreateAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(entities.ErrNotFound, "user not found")
		}
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	teamsID, err := s.getIDs(ctx, query2, user.ID)
	if err != nil {
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(err, "get teams ID")
	}

	user.TeamsID = teamsID

	return user, nil
}

func (s *Storage) CreateUser(ctx context.Context, name, email, hashPass string) (*entities.User, error) {
	slog.Info("CreateUser")
	query := `INSERT INTO users (name, email, hash_pass, create_at) 
				VALUES ($1, $2, $3, NOW())
				RETURNING id, create_at`

	var (
		newUserID    int64
		userCreateAt time.Time
	)

	err := s.db.QueryRow(ctx, query, name, email, hashPass).Scan(&newUserID, &userCreateAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(entities.ErrInternal, "create user")
		}
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	newUser, err := entities.NewUser(newUserID, name, email, []byte(hashPass), userCreateAt)
	if err != nil {
		slog.Error("Create User", "err", err)
		return nil, errors.Wrap(err, "create user")
	}

	return newUser, nil
}

func (s *Storage) AddMemberInTeam(ctx context.Context, userID, teamID int64) error {
	slog.Info("AddMemberInTeam")
	query := `INSERT INTO teams_users (user_id, team_id) VALUES ($1, $2)`
	commandTag, err := s.db.Exec(ctx, query, userID, teamID)
	if err != nil {
		slog.Error("Query", "err", err)
		return errors.Wrap(entities.ErrInternal, err.Error())
	}
	if commandTag.RowsAffected() == 0 {
		slog.Error("Query", "err", "teams_users already exists")
		return errors.Wrap(entities.ErrInternal, "insert user")
	}

	return nil
}

func (s *Storage) GetTeam(ctx context.Context, teamID int64) (*entities.Team, error) {
	slog.Info("GetTeam")
	query := `SELECT team_id, name, owner_id, create_at FROM teams WHERE team_id = $1`
	query2 := `SELECT user_id FROM teams_users WHERE team_id = $1`
	query3 := `SELECT id FROM tasks WHERE team_id = $1`

	team := &entities.Team{}

	err := s.db.QueryRow(ctx, query, teamID).Scan(&team.ID, &team.Name, &team.OwnerID, &team.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(entities.ErrNotFound, "team not found")
		}
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	membersID, err := s.getIDs(ctx, query2, team.ID)
	if err != nil {
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(err, "get team members")
	}

	team.MembersID = membersID

	tasksID, err := s.getIDs(ctx, query3, team.ID)
	if err != nil {
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(err, "get team tasks")
	}

	team.TasksID = tasksID

	return team, nil
}

func (s *Storage) CreateTeam(ctx context.Context, userID int64, name string) (*entities.Team, error) {
	slog.Info("CreateTeam")
	query := `INSERT INTO teams (name, owner_id, create_at) 
				VALUES ($1, $2, NOW()) 
				RETURNING team_id, create_at`

	var (
		newTeamID     int64
		teamCreatedAt time.Time
	)

	err := s.db.QueryRow(ctx, query, name, userID).Scan(&newTeamID, &teamCreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(entities.ErrInternal, "create team")
		}
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	newTeam, err := entities.NewTeam(newTeamID, name, userID, teamCreatedAt)
	if err != nil {
		slog.Error("Create Team", "err", err)
		return nil, errors.Wrap(err, "create team")
	}

	return newTeam, nil
}

func (s *Storage) GetAllTeams(ctx context.Context, userID int64) ([]*entities.Team, error) {
	slog.Info("GetAllTeams")
	query := `SELECT t.team_id, t.name, t.owner_id, t.create_at 
          		FROM teams t 
          		JOIN teams_users tu ON t.team_id = tu.team_id 
          		WHERE tu.user_id = $1`
	query2 := `SELECT user_id FROM teams_users WHERE team_id = $1`
	query3 := `SELECT id FROM tasks WHERE team_id = $1`

	teams := make([]*entities.Team, 0)

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var (
			teamID        int64
			teamName      string
			teamOwnerID   int64
			teamCreatedAt time.Time
		)

		err := rows.Scan(&teamID, &teamName, &teamOwnerID, &teamCreatedAt)
		if err != nil {
			slog.Error("Scan", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}

		team, err := entities.NewTeam(teamID, teamName, teamOwnerID, teamCreatedAt)
		if err != nil {
			slog.Error("Create Team", "err", err)
			return nil, errors.Wrap(err, "create team")
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Rows.Err", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	for _, team := range teams {
		membersID, err := s.getIDs(ctx, query2, team.ID)
		if err != nil {
			slog.Error("Query", "err", err)
			return nil, errors.Wrap(err, "get team members")
		}

		tasksID, err := s.getIDs(ctx, query3, team.ID)
		if err != nil {
			slog.Error("Query", "err", err)
			return nil, errors.Wrap(err, "get team tasks")
		}

		team.MembersID = membersID
		team.TasksID = tasksID
	}

	return teams, nil
}

func (s *Storage) CreateTask(ctx context.Context, userID, assigneeID, teamID int64, title, description string) (*entities.Task, error) {
	slog.Info("CreateTask")
	query := `INSERT INTO tasks (assignee_id, owner_id, team_id, title, description, status, create_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())
				RETURNING task_id, create_at`

	var (
		newTaskID int64
		createAt  time.Time
	)

	err := s.db.QueryRow(ctx, query, assigneeID, userID, teamID, title, description, entities.StatusCreated).Scan(&newTaskID, &createAt)
	if err != nil {
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	newTask, err := entities.NewTask(newTaskID, userID, teamID, assigneeID, title, description, entities.StatusCreated, createAt)
	if err != nil {
		slog.Error("Create Task", "err", err)
		return nil, errors.Wrap(err, "create task")
	}

	return newTask, nil
}

func (s *Storage) GetTask(ctx context.Context, taskID int64) (*entities.Task, error) {
	slog.Info("GetTask")
	query := `SELECT assignee_id, owner_id, team_id, title, description, status, create_at
				FROM tasks
				WHERE id = $1`
	query2 := `SELECT user_id, comm, create_at
				FROM comments
				WHERE task_id = $1`

	var (
		assigneeID, ownerID, teamID int64
		title, description, status  string
		createAt                    time.Time
	)

	err := s.db.QueryRow(ctx, query, taskID).Scan(&assigneeID, &ownerID, &teamID, &title, &description, &status, &createAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Wrap(entities.ErrNotFound, "task not found")
		}
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	task, err := entities.NewTask(taskID, ownerID, teamID, assigneeID, title, description, status, createAt)
	if err != nil {
		slog.Error("Create Task", "err", err)
		return nil, errors.Wrap(err, "get task")
	}

	rows, err := s.db.Query(ctx, query2, taskID)
	if err != nil {
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userID   int64
			comm     string
			createAt time.Time
		)

		err := rows.Scan(&userID, &comm, &createAt)
		if err != nil {
			slog.Error("Scan", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}

		comment, err := entities.NewComment(userID, comm)
		if err != nil {
			slog.Error("Create Comment", "err", err)
			return nil, errors.Wrap(err, "create comment")
		}

		comment.SetCreateAt(createAt)

		err = task.AddComment(comment)
		if err != nil {
			slog.Error("Add Comment", "err", err)
			return nil, errors.Wrap(err, "add comment")
		}
	}

	if err := rows.Err(); err != nil {
		slog.Error("Rows.Err", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	return task, nil
}

func (s *Storage) UpdateTask(ctx context.Context, task *entities.Task) error {
	slog.Info("UpdateTask")
	query := `UPDATE tasks
				SET assignee_id = $1, title = $2, status = $3, description = $4
				WHERE id = $5`

	commandTag, err := s.db.Exec(ctx, query, task.AssigneeID, task.Title, task.Status, task.Description, task.ID)
	if err != nil {
		slog.Error("Query", "err", err)
		return errors.Wrap(entities.ErrInternal, err.Error())
	}
	if commandTag.RowsAffected() == 0 {
		slog.Error("Query", "err", "no rows updated")
		return errors.Wrap(entities.ErrNotFound, "task not found")
	}

	return nil
}

func (s *Storage) AddComment(ctx context.Context, taskID int64, comment *entities.Comment) error {
	slog.Info("AddComment")
	query := `INSERT INTO comments (user_id, task_id, comm, create_at)
				VALUES ($1, $2, $3, NOW())
				RETURNING create_at`

	var createAt time.Time

	err := s.db.QueryRow(ctx, query, comment.UserID, taskID, comment.Comm).Scan(&createAt)
	if err != nil {
		slog.Error("Query", "err", err)
		return errors.Wrap(entities.ErrInternal, err.Error())
	}

	comment.SetCreateAt(createAt)
	return nil
}

func (s *Storage) GetHistory(ctx context.Context, taskID int64) (*entities.History, error) {
	slog.Info("GetHistory")
	query := `SELECT user_id, status, title, description, create_at
				FROM records
				WHERE task_id = $1`

	rows, err := s.db.Query(ctx, query, taskID)
	if err != nil {
		slog.Error("Query", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}
	defer rows.Close()

	history := entities.NewHistory()

	for rows.Next() {
		var (
			userID      int64
			status      string
			title       string
			description string
			createdAt   time.Time
		)

		err = rows.Scan(&userID, &status, &title, &description, &createdAt)
		if err != nil {
			slog.Error("Scan", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}

		record, err := entities.NewRecord(userID, taskID, status, title, description, createdAt)
		if err != nil {
			slog.Error("Create Record", "err", err)
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}

		history.AddRecord(record)
	}
	if err := rows.Err(); err != nil {
		slog.Error("Rows.Err", "err", err)
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	return history, nil
}

func (s *Storage) AddHistory(ctx context.Context, record *entities.Record) error {
	slog.Info("AddHistory")
	query := `INSERT INTO records (user_id, task_id, status, title, description, create_at)
				VALUES ($1, $2, $3, $4, $5, NOW())
				RETURNING create_at`

	var createAt time.Time

	err := s.db.QueryRow(ctx, query, record.GetUserID(), record.GetTaskID(), record.GetStatus(), record.GetTitle(), record.GetDescription()).Scan(&createAt)
	if err != nil {
		slog.Error("Query", "err", err)
		return errors.Wrap(entities.ErrInternal, err.Error())
	}
	record.SetCreateAt(createAt)
	return nil
}

func (s *Storage) getIDs(ctx context.Context, query string, args ...interface{}) ([]int64, error) {
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	defer rows.Close()

	ids := make([]int64, 0)

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(entities.ErrInternal, err.Error())
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(entities.ErrInternal, err.Error())
	}

	return ids, nil
}
