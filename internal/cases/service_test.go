// internal/cases/service_test.go
package cases_test

import (
	"context"
	"test_task/internal/cases"
	"test_task/internal/cases/mocks"
	"test_task/internal/entities"
	"test_task/pkg/dto"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

var (
	errTest = errors.New("test error")
)

// ========== MOCK CONFIG ==========
type mockConfig struct{}

func (m *mockConfig) TokenSecret() string                     { return "test-secret" }
func (m *mockConfig) TokenDuration() time.Duration            { return 24 * time.Hour }
func (m *mockConfig) HTTPPort() string                        { return "8080" }
func (m *mockConfig) HTTPTimeout() time.Duration              { return 10 * time.Second }
func (m *mockConfig) MetricsTimeout() time.Duration           { return 10 * time.Second }
func (m *mockConfig) MetricsPort() string                     { return "9090" }
func (m *mockConfig) AddSource() bool                         { return true }
func (m *mockConfig) LoggerLevel() string                     { return "info" }
func (m *mockConfig) StorageConnStr() string                  { return "" }
func (m *mockConfig) GracefullShutdownTimeout() time.Duration { return 5 * time.Second }
func (m *mockConfig) LifeIdleConns() time.Duration            { return 5 * time.Minute }
func (m *mockConfig) MaxOpenConns() int                       { return 25 }
func (m *mockConfig) MaxIdleConns() int                       { return 10 }
func (m *mockConfig) LifeConns() time.Duration                { return 5 * time.Minute }
func (m *mockConfig) CacheHost() string                       { return "localhost" }
func (m *mockConfig) CachePort() string                       { return "6379" }
func (m *mockConfig) CacheTTL() time.Duration                 { return 5 * time.Minute }
func (m *mockConfig) CachePass() string                       { return "" }
func (m *mockConfig) CacheDB() int                            { return 0 }
func (m *mockConfig) RateLimiterRate() float64                { return 100 }
func (m *mockConfig) RateLimiterBurst() int                   { return 100 }

func TestService_CreateUser_Success(t *testing.T) {
	// 1. Создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 2. Создаём моки
	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	// 3. Создаём сервис
	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	name := "John Doe"
	email := "john@example.com"
	password := "password123"

	// 4. Настраиваем моки
	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(nil, entities.ErrNotFound)

	expectedUser := &entities.User{ID: 1, Name: name, Email: email}
	mockRepo.EXPECT().
		CreateUser(ctx, name, email, gomock.Any()).
		Return(expectedUser, nil)

	// 5. Вызываем метод
	user, err := svc.CreateUser(ctx, name, email, password)

	// 6. Проверяем
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, int64(1), user.ID)
	require.Equal(t, name, user.Name)
	require.Equal(t, email, user.Email)
}

func TestService_CreateUser_EmailAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	name := "John Doe"
	email := "john@example.com"
	password := "password123"

	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(&entities.User{Email: email}, nil)

	user, err := svc.CreateUser(ctx, name, email, password)

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "email is exist")
}

func TestService_CreateUser_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	user, err := svc.CreateUser(ctx, "", "john@example.com", "password123")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "name is empty")
}

func TestService_CreateUser_GetUserByEmailError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	name := "John Doe"
	email := "john@example.com"
	password := "password123"

	// ❌ GetUserByEmail возвращает ошибку (не ErrNotFound)
	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(nil, errTest)

	user, err := svc.CreateUser(ctx, name, email, password)

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "failed to check user existence")
}

func TestService_CreateUser_CreateUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	name := "John Doe"
	email := "john@example.com"
	password := "password123"

	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(nil, entities.ErrNotFound)

	mockRepo.EXPECT().
		CreateUser(ctx, name, email, gomock.Any()).
		Return(nil, errTest)

	user, err := svc.CreateUser(ctx, name, email, password)

	require.Error(t, err)
	require.Nil(t, user)
	require.ErrorIs(t, err, errTest)
}

func TestService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	email := "john@example.com"
	password := "password123"

	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(&entities.User{ID: 1, Email: email, HashPassword: hashPass}, nil)

	user, token, err := svc.Login(ctx, email, password)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.NotEmpty(t, token)
}

func TestService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	email := "notfound@example.com"
	password := "password123"

	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(nil, entities.ErrNotFound)

	user, token, err := svc.Login(ctx, email, password)

	require.Error(t, err)
	require.Nil(t, user)
	require.Empty(t, token)
}

func TestService_Login_CheckHashError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	email := "john@example.com"
	password := "password123"

	// ❌ Возвращаем пользователя с кривым хешем, чтобы CheckHashPass упал
	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(&entities.User{
			ID:           1,
			Email:        email,
			HashPassword: []byte("invalid-hash"),
		}, nil)

	user, token, err := svc.Login(ctx, email, password)

	require.Error(t, err)
	require.Nil(t, user)
	require.Empty(t, token)
	require.Contains(t, err.Error(), "check hash")
}

func TestService_GetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	expectedUser := &entities.User{ID: userID, Name: "John Doe", Email: "john@example.com"}

	mockRepo.EXPECT().
		GetUserByID(ctx, userID).
		Return(expectedUser, nil)

	user, err := svc.GetUserByID(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, expectedUser.ID, user.ID)
	require.Equal(t, expectedUser.Name, user.Name)
	require.Equal(t, expectedUser.Email, user.Email)
}

func TestService_GetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(999)

	mockRepo.EXPECT().
		GetUserByID(ctx, userID).
		Return(nil, entities.ErrNotFound)

	user, err := svc.GetUserByID(ctx, userID)

	require.Error(t, err)
	require.Nil(t, user)
	require.ErrorIs(t, err, entities.ErrNotFound)
	require.Contains(t, err.Error(), "user not found")
}

func TestService_GetUserByID_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(0)

	user, err := svc.GetUserByID(ctx, userID)

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "invalid id")
}

func TestService_GetUserByID_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)

	mockRepo.EXPECT().
		GetUserByID(ctx, userID).
		Return(nil, errTest)

	user, err := svc.GetUserByID(ctx, userID)

	require.Error(t, err)
	require.Nil(t, user)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTaskByID_FromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	expectedTask := &entities.Task{ID: taskID, Title: "Test Task"}

	mockCache.EXPECT().
		GetTask(ctx, taskID).
		Return(expectedTask, nil)

	task, err := svc.GetTaskByID(ctx, taskID)

	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, expectedTask.ID, task.ID)
}

func TestService_GetTaskByID_CacheMiss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	expectedTask := &entities.Task{ID: taskID, Title: "Test Task"}

	mockCache.EXPECT().
		GetTask(ctx, taskID).
		Return(nil, nil)

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(expectedTask, nil)

	mockCache.EXPECT().
		SetTask(ctx, expectedTask).
		Return(nil)

	task, err := svc.GetTaskByID(ctx, taskID)

	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, expectedTask.ID, task.ID)
}

func TestService_CreateTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamName := "Team A"

	mockRepo.EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, cases.Repository) error) error {
			return fn(ctx, mockRepo)
		})

	mockRepo.EXPECT().
		CreateTeam(ctx, teamName, userID).
		Return(&entities.Team{ID: 1, Name: teamName, OwnerID: userID}, nil)

	mockRepo.EXPECT().
		AddMember(ctx, userID, int64(1), "owner").
		Return(nil)

	team, err := svc.CreateTeam(ctx, userID, teamName)

	require.NoError(t, err)
	require.NotNil(t, team)
	require.Equal(t, teamName, team.Name)
}

func TestService_CreateTeam_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	team, err := svc.CreateTeam(ctx, 1, "")

	require.Error(t, err)
	require.Nil(t, team)
	require.Contains(t, err.Error(), "name is empty")
}

func TestService_GetUserByEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	email := "john@example.com"
	expectedUser := &entities.User{ID: 1, Email: email, Name: "John Doe"}

	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(expectedUser, nil)

	user, err := svc.GetUserByEmail(ctx, email)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, expectedUser.ID, user.ID)
	require.Equal(t, expectedUser.Email, user.Email)
	require.Equal(t, expectedUser.Name, user.Name)
}

func TestService_GetUserByEmail_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	email := "notfound@example.com"

	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(nil, entities.ErrNotFound)

	user, err := svc.GetUserByEmail(ctx, email)

	require.Error(t, err)
	require.Nil(t, user)
	require.ErrorIs(t, err, entities.ErrNotFound)
}

func TestService_GetUserByEmail_EmptyEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	user, err := svc.GetUserByEmail(ctx, "")

	require.Error(t, err)
	require.Nil(t, user)
	require.Contains(t, err.Error(), "email is empty")
}

func TestService_GetUserByEmail_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	email := "john@example.com"

	mockRepo.EXPECT().
		GetUserByEmail(ctx, email).
		Return(nil, errTest)

	user, err := svc.GetUserByEmail(ctx, email)

	require.Error(t, err)
	require.Nil(t, user)
	require.ErrorIs(t, err, errTest)
}

func TestService_AddMember_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1) // приглашающий (admin)
	teamID := int64(1)
	memberID := int64(2) // приглашаемый
	role := "member"

	// 1. Проверяем, что команда существует
	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(&entities.Team{ID: teamID, Name: "Team A"}, nil)

	// 2. Проверяем, что приглашающий — admin или owner
	mockRepo.EXPECT().
		IsAdminOrOwner(ctx, userID, teamID).
		Return(true, nil)

	// 3. Проверяем, что приглашаемый пользователь существует
	mockRepo.EXPECT().
		GetUserByID(ctx, memberID).
		Return(&entities.User{ID: memberID, Name: "Jane Doe"}, nil)

	// 4. Добавляем участника
	mockRepo.EXPECT().
		AddMember(ctx, memberID, teamID, role).
		Return(nil)

	err = svc.AddMember(ctx, userID, teamID, memberID, role)

	require.NoError(t, err)
}

func TestService_AddMember_NotAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(3) // обычный участник (не admin)
	teamID := int64(1)
	memberID := int64(2)
	role := "member"

	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(&entities.Team{ID: teamID, Name: "Team A"}, nil)

	mockRepo.EXPECT().
		IsAdminOrOwner(ctx, userID, teamID).
		Return(false, nil)

	err = svc.AddMember(ctx, userID, teamID, memberID, role)

	require.Error(t, err)
	require.Contains(t, err.Error(), "user is not an admin")
}

func TestService_AddMember_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(999)
	memberID := int64(2)
	role := "member"

	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(nil, entities.ErrNotFound)

	err = svc.AddMember(ctx, userID, teamID, memberID, role)

	require.Error(t, err)
	require.ErrorIs(t, err, entities.ErrNotFound)
}

func TestService_AddMember_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(1)
	memberID := int64(999)
	role := "member"

	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(&entities.Team{ID: teamID, Name: "Team A"}, nil)

	mockRepo.EXPECT().
		IsAdminOrOwner(ctx, userID, teamID).
		Return(true, nil)

	mockRepo.EXPECT().
		GetUserByID(ctx, memberID).
		Return(nil, entities.ErrNotFound)

	err = svc.AddMember(ctx, userID, teamID, memberID, role)

	require.Error(t, err)
	require.ErrorIs(t, err, entities.ErrNotFound)
}

func TestService_AddMember_AlreadyInTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(1)
	memberID := int64(2)
	role := "member"

	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(&entities.Team{ID: teamID, Name: "Team A"}, nil)

	mockRepo.EXPECT().
		IsAdminOrOwner(ctx, userID, teamID).
		Return(true, nil)

	mockRepo.EXPECT().
		GetUserByID(ctx, memberID).
		Return(&entities.User{ID: memberID, Name: "Jane Doe"}, nil)

	mockRepo.EXPECT().
		AddMember(ctx, memberID, teamID, role).
		Return(entities.ErrConflict)

	err = svc.AddMember(ctx, userID, teamID, memberID, role)

	require.Error(t, err)
	require.Contains(t, err.Error(), "user already in team")
}

func TestService_AddMember_InvalidMemberID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(1)
	memberID := int64(0)
	role := "member"

	err = svc.AddMember(ctx, userID, teamID, memberID, role)

	require.Error(t, err)
	require.Contains(t, err.Error(), "memberID is invalid")
}

func TestService_AddMember_DefaultRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(1)
	memberID := int64(2)
	role := "" // пустая роль → должна стать "member"

	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(&entities.Team{ID: teamID, Name: "Team A"}, nil)

	mockRepo.EXPECT().
		IsAdminOrOwner(ctx, userID, teamID).
		Return(true, nil)

	mockRepo.EXPECT().
		GetUserByID(ctx, memberID).
		Return(&entities.User{ID: memberID, Name: "Jane Doe"}, nil)

	// Ожидаем, что роль будет "member"
	mockRepo.EXPECT().
		AddMember(ctx, memberID, teamID, "member").
		Return(nil)

	err = svc.AddMember(ctx, userID, teamID, memberID, role)

	require.NoError(t, err)
}

func TestService_GetTeams_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamIDs := []int64{1, 2}
	teams := []*entities.Team{
		{ID: 1, Name: "Team A", OwnerID: 1},
		{ID: 2, Name: "Team B", OwnerID: 2},
	}

	mockRepo.EXPECT().
		GetUserTeams(ctx, userID).
		Return(teamIDs, nil)

	mockRepo.EXPECT().
		GetTeamsByIDs(ctx, teamIDs).
		Return(teams, nil)

	mockRepo.EXPECT().
		GetMemberRole(ctx, userID, int64(1)).
		Return("owner", nil)

	mockRepo.EXPECT().
		GetMemberRole(ctx, userID, int64(2)).
		Return("member", nil)

	result, err := svc.GetTeams(ctx, userID)

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "owner", result[0].Role)
	require.Equal(t, "member", result[1].Role)
	require.Equal(t, teams[0].ID, result[0].Team.ID)
	require.Equal(t, teams[0].Name, result[0].Team.Name)
}

func TestService_GetTeams_NoTeams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)

	mockRepo.EXPECT().
		GetUserTeams(ctx, userID).
		Return([]int64{}, nil)

	result, err := svc.GetTeams(ctx, userID)

	require.NoError(t, err)
	require.Empty(t, result)
}

func TestService_GetTeams_GetUserTeamsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)

	mockRepo.EXPECT().
		GetUserTeams(ctx, userID).
		Return(nil, errTest)

	result, err := svc.GetTeams(ctx, userID)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTeams_GetTeamsByIDsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamIDs := []int64{1, 2}

	mockRepo.EXPECT().
		GetUserTeams(ctx, userID).
		Return(teamIDs, nil)

	mockRepo.EXPECT().
		GetTeamsByIDs(ctx, teamIDs).
		Return(nil, errTest)

	result, err := svc.GetTeams(ctx, userID)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTeams_GetMemberRoleError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamIDs := []int64{1, 2}
	teams := []*entities.Team{
		{ID: 1, Name: "Team A", OwnerID: 1},
		{ID: 2, Name: "Team B", OwnerID: 2},
	}

	mockRepo.EXPECT().
		GetUserTeams(ctx, userID).
		Return(teamIDs, nil)

	mockRepo.EXPECT().
		GetTeamsByIDs(ctx, teamIDs).
		Return(teams, nil)

	mockRepo.EXPECT().
		GetMemberRole(ctx, userID, int64(1)).
		Return("", errTest)

	result, err := svc.GetTeams(ctx, userID)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTeamByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)
	expectedTeam := &entities.Team{
		ID:      1,
		Name:    "Team A",
		OwnerID: 1,
	}

	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(expectedTeam, nil)

	team, err := svc.GetTeamByID(ctx, teamID)

	require.NoError(t, err)
	require.NotNil(t, team)
	require.Equal(t, expectedTeam.ID, team.ID)
	require.Equal(t, expectedTeam.Name, team.Name)
	require.Equal(t, expectedTeam.OwnerID, team.OwnerID)
}

func TestService_GetTeamByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(999)

	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(nil, entities.ErrNotFound)

	team, err := svc.GetTeamByID(ctx, teamID)

	require.Error(t, err)
	require.Nil(t, team)
	require.ErrorIs(t, err, entities.ErrNotFound)
}

func TestService_GetTeamByID_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)

	mockRepo.EXPECT().
		GetTeamByID(ctx, teamID).
		Return(nil, errTest)

	team, err := svc.GetTeamByID(ctx, teamID)

	require.Error(t, err)
	require.Nil(t, team)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTeamMembers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)
	expectedMembers := []*entities.TeamMember{
		{UserID: 1, TeamID: 1, Role: "owner"},
		{UserID: 2, TeamID: 1, Role: "member"},
		{UserID: 3, TeamID: 1, Role: "admin"},
	}

	mockRepo.EXPECT().
		GetTeamMembers(ctx, teamID).
		Return(expectedMembers, nil)

	members, err := svc.GetTeamMembers(ctx, teamID)

	require.NoError(t, err)
	require.NotNil(t, members)
	require.Len(t, members, 3)
	require.Equal(t, expectedMembers[0].UserID, members[0].UserID)
	require.Equal(t, expectedMembers[0].Role, members[0].Role)
	require.Equal(t, expectedMembers[1].UserID, members[1].UserID)
	require.Equal(t, expectedMembers[1].Role, members[1].Role)
}

func TestService_GetTeamMembers_EmptyList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)

	mockRepo.EXPECT().
		GetTeamMembers(ctx, teamID).
		Return([]*entities.TeamMember{}, nil)

	members, err := svc.GetTeamMembers(ctx, teamID)

	require.NoError(t, err)
	require.Empty(t, members)
}

func TestService_GetTeamMembers_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(999)

	mockRepo.EXPECT().
		GetTeamMembers(ctx, teamID).
		Return(nil, entities.ErrNotFound)

	members, err := svc.GetTeamMembers(ctx, teamID)

	require.Error(t, err)
	require.Nil(t, members)
	require.ErrorIs(t, err, entities.ErrNotFound)
}

func TestService_GetTeamMembers_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)

	mockRepo.EXPECT().
		GetTeamMembers(ctx, teamID).
		Return(nil, errTest)

	members, err := svc.GetTeamMembers(ctx, teamID)

	require.Error(t, err)
	require.Nil(t, members)
	require.ErrorIs(t, err, errTest)
}

func TestService_CreateTask_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl) // ← Cache для SetTask и DeleteTasksByTeam
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	assigneeID := int64(2)
	teamID := int64(1)
	title := "Test Task"
	description := "Test Description"
	expectedTask := &entities.Task{ID: 1, Title: title, Description: description}

	mockRepo.EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, cases.Repository) error) error {
			return fn(ctx, mockRepo)
		})

	mockRepo.EXPECT().
		CreateTask(ctx, userID, assigneeID, teamID, title, description, entities.TaskStatusTODO).
		Return(expectedTask, nil)

	mockRepo.EXPECT().
		AddHistoryRecord(ctx, gomock.Any()).
		Return(nil)

	mockCache.EXPECT().
		SetTask(ctx, expectedTask).
		Return(nil)

	mockCache.EXPECT().
		DeleteTasksByTeam(ctx, teamID).
		Return(nil)

	task, err := svc.CreateTask(ctx, userID, assigneeID, teamID, title, description)

	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, expectedTask.ID, task.ID)
	require.Equal(t, title, task.Title)
	require.Equal(t, description, task.Description)
}

func TestService_CreateTask_EmptyTitle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	assigneeID := int64(2)
	teamID := int64(1)
	title := ""
	description := "Test Description"

	task, err := svc.CreateTask(ctx, userID, assigneeID, teamID, title, description)

	require.Error(t, err)
	require.Nil(t, task)
	require.Contains(t, err.Error(), "title is empty")
}

func TestService_CreateTask_EmptyDescription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	assigneeID := int64(2)
	teamID := int64(1)
	title := "Test Task"
	description := ""

	task, err := svc.CreateTask(ctx, userID, assigneeID, teamID, title, description)

	require.Error(t, err)
	require.Nil(t, task)
	require.Contains(t, err.Error(), "description is empty")
}

func TestService_CreateTask_InvalidAssigneeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	assigneeID := int64(0)
	teamID := int64(1)
	title := "Test Task"
	description := "Test Description"

	task, err := svc.CreateTask(ctx, userID, assigneeID, teamID, title, description)

	require.Error(t, err)
	require.Nil(t, task)
	require.Contains(t, err.Error(), "assigneeID is invalid")
}

func TestService_CreateTask_InvalidTeamID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	assigneeID := int64(2)
	teamID := int64(0)
	title := "Test Task"
	description := "Test Description"

	task, err := svc.CreateTask(ctx, userID, assigneeID, teamID, title, description)

	require.Error(t, err)
	require.Nil(t, task)
	require.Contains(t, err.Error(), "teamID is invalid")
}

func TestService_CreateTask_CreateTaskError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	assigneeID := int64(2)
	teamID := int64(1)
	title := "Test Task"
	description := "Test Description"

	mockRepo.EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, cases.Repository) error) error {
			return fn(ctx, mockRepo)
		})

	mockRepo.EXPECT().
		CreateTask(ctx, userID, assigneeID, teamID, title, description, entities.TaskStatusTODO).
		Return(nil, errTest)

	task, err := svc.CreateTask(ctx, userID, assigneeID, teamID, title, description)

	require.Error(t, err)
	require.Nil(t, task)
	require.ErrorIs(t, err, errTest)
}

func TestService_CreateTask_AddHistoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	assigneeID := int64(2)
	teamID := int64(1)
	title := "Test Task"
	description := "Test Description"
	expectedTask := &entities.Task{ID: 1, Title: title, Description: description}

	mockRepo.EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, cases.Repository) error) error {
			return fn(ctx, mockRepo)
		})

	mockRepo.EXPECT().
		CreateTask(ctx, userID, assigneeID, teamID, title, description, entities.TaskStatusTODO).
		Return(expectedTask, nil)

	mockRepo.EXPECT().
		AddHistoryRecord(ctx, gomock.Any()).
		Return(errTest)

	task, err := svc.CreateTask(ctx, userID, assigneeID, teamID, title, description)

	require.Error(t, err)
	require.Nil(t, task)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTaskByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	expectedTask := &entities.Task{
		ID:    taskID,
		Title: "Test Task",
	}

	mockCache.EXPECT().
		GetTask(ctx, taskID).
		Return(expectedTask, nil)

	task, err := svc.GetTaskByID(ctx, taskID)

	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, expectedTask.ID, task.ID)
	require.Equal(t, expectedTask.Title, task.Title)
}

func TestService_GetTaskByID_CacheMiss_DBSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	expectedTask := &entities.Task{
		ID:    taskID,
		Title: "Test Task",
	}

	mockCache.EXPECT().
		GetTask(ctx, taskID).
		Return(nil, nil)

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(expectedTask, nil)

	mockCache.EXPECT().
		SetTask(ctx, expectedTask).
		Return(nil)

	task, err := svc.GetTaskByID(ctx, taskID)

	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, expectedTask.ID, task.ID)
	require.Equal(t, expectedTask.Title, task.Title)
}

func TestService_GetTaskByID_CacheMiss_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)

	mockCache.EXPECT().
		GetTask(ctx, taskID).
		Return(nil, nil)

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(nil, errTest)

	task, err := svc.GetTaskByID(ctx, taskID)

	require.Error(t, err)
	require.Nil(t, task)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTaskByID_CacheGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)

	// Кеш вернул ошибку (не критично, идём в БД)
	mockCache.EXPECT().
		GetTask(ctx, taskID).
		Return(nil, errTest)

	expectedTask := &entities.Task{
		ID:    taskID,
		Title: "Test Task",
	}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(expectedTask, nil)

	mockCache.EXPECT().
		SetTask(ctx, expectedTask).
		Return(nil)

	task, err := svc.GetTaskByID(ctx, taskID)

	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, expectedTask.ID, task.ID)
}

func TestService_GetTaskByID_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(0)

	task, err := svc.GetTaskByID(ctx, taskID)

	require.Error(t, err)
	require.Nil(t, task)
	require.Contains(t, err.Error(), "taskID is invalid")
}

func TestService_GetTasksByTeam_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)
	limit := 20
	offset := 0
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: teamID},
		{ID: 2, Title: "Task 2", TeamID: teamID},
	}

	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(expectedTasks, nil)

	tasks, err := svc.GetTasksByTeam(ctx, teamID, limit, offset)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
	require.Equal(t, expectedTasks[0].ID, tasks[0].ID)
	require.Equal(t, expectedTasks[0].Title, tasks[0].Title)
}

func TestService_GetTasksByTeam_CacheMiss_DBSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)
	limit := 20
	offset := 0
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: teamID},
		{ID: 2, Title: "Task 2", TeamID: teamID},
	}

	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(nil, nil)

	mockRepo.EXPECT().
		GetTasksByTeam(ctx, teamID, 0, 0).
		Return(expectedTasks, nil)

	mockCache.EXPECT().
		SetTasksByTeam(ctx, teamID, expectedTasks).
		Return(nil)

	tasks, err := svc.GetTasksByTeam(ctx, teamID, limit, offset)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
	require.Equal(t, expectedTasks[0].ID, tasks[0].ID)
}

func TestService_GetTasksByTeam_CacheMiss_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)
	limit := 20
	offset := 0

	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(nil, nil)

	mockRepo.EXPECT().
		GetTasksByTeam(ctx, teamID, 0, 0).
		Return(nil, errTest)

	tasks, err := svc.GetTasksByTeam(ctx, teamID, limit, offset)

	require.Error(t, err)
	require.Nil(t, tasks)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTasksByTeam_CacheGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)
	limit := 20
	offset := 0
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: teamID},
		{ID: 2, Title: "Task 2", TeamID: teamID},
	}

	// Кеш вернул ошибку (не критично, идём в БД)
	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(nil, errTest)

	mockRepo.EXPECT().
		GetTasksByTeam(ctx, teamID, 0, 0).
		Return(expectedTasks, nil)

	mockCache.EXPECT().
		SetTasksByTeam(ctx, teamID, expectedTasks).
		Return(nil)

	tasks, err := svc.GetTasksByTeam(ctx, teamID, limit, offset)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
}

func TestService_GetTasksByTeam_EmptyTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(1)
	limit := 20
	offset := 0

	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(nil, nil)

	mockRepo.EXPECT().
		GetTasksByTeam(ctx, teamID, 0, 0).
		Return([]*entities.Task{}, nil)

	tasks, err := svc.GetTasksByTeam(ctx, teamID, limit, offset)

	require.NoError(t, err)
	require.Empty(t, tasks)
}

func TestService_GetTasksByTeam_InvalidTeamID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	teamID := int64(0)
	limit := 20
	offset := 0

	// Метод не проверяет teamID, но передаёт в репозиторий
	// Тест проверяет, что ошибка будет от репозитория
	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(nil, nil)

	mockRepo.EXPECT().
		GetTasksByTeam(ctx, teamID, 0, 0).
		Return(nil, errTest)

	tasks, err := svc.GetTasksByTeam(ctx, teamID, limit, offset)

	require.Error(t, err)
	require.Nil(t, tasks)
}

func TestService_UpdateTask_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	teamID := int64(1)
	oldTask := &entities.Task{
		ID:          taskID,
		Title:       "Old Title",
		Description: "Old Description",
		Status:      entities.TaskStatusTODO,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    time.Now(),
	}
	updatedTask := &entities.Task{
		ID:          taskID,
		Title:       "New Title",
		Description: "New Description",
		Status:      entities.TaskStatusInProgress,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    oldTask.CreateAt,
	}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(oldTask, nil)

	mockRepo.EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, cases.Repository) error) error {
			return fn(ctx, mockRepo)
		})

	mockRepo.EXPECT().
		UpdateTask(ctx, updatedTask).
		Return(nil)

	// ✅ AddHistoryRecord вызывается 3 раза (для title, description, status)
	mockRepo.EXPECT().
		AddHistoryRecord(ctx, gomock.Any()).
		Return(nil).
		Times(3) // ← важно: 3 раза!

	mockCache.EXPECT().
		SetTask(ctx, updatedTask).
		Return(nil)

	mockCache.EXPECT().
		DeleteTasksByTeam(ctx, teamID).
		Return(nil)

	err = svc.UpdateTask(ctx, userID, updatedTask)

	require.NoError(t, err)
}

func TestService_UpdateTask_NoChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	teamID := int64(1)
	oldTask := &entities.Task{
		ID:          taskID,
		Title:       "Same Title",
		Description: "Same Description",
		Status:      entities.TaskStatusTODO,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    time.Now(),
	}
	updatedTask := &entities.Task{
		ID:          taskID,
		Title:       "Same Title",
		Description: "Same Description",
		Status:      entities.TaskStatusTODO,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    oldTask.CreateAt,
	}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(oldTask, nil)

	// ❌ Нет изменений → InTx НЕ вызывается
	// ❌ UpdateTask НЕ вызывается
	// ❌ AddHistoryRecord НЕ вызывается
	// ❌ SetTask НЕ вызывается
	// ❌ DeleteTasksByTeam НЕ вызывается

	err = svc.UpdateTask(ctx, userID, updatedTask)

	require.NoError(t, err)
}

func TestService_UpdateTask_TaskNil(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)

	err = svc.UpdateTask(ctx, userID, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "task is empty")
}

func TestService_UpdateTask_InvalidUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(0)
	task := &entities.Task{ID: 1}

	err = svc.UpdateTask(ctx, userID, task)

	require.Error(t, err)
	require.Contains(t, err.Error(), "userID is invalid")
}

func TestService_UpdateTask_TaskNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(999)
	task := &entities.Task{ID: taskID, Title: "Task"}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(nil, entities.ErrNotFound)

	err = svc.UpdateTask(ctx, userID, task)

	require.Error(t, err)
	require.ErrorIs(t, err, entities.ErrNotFound)
}

func TestService_UpdateTask_ChangeAssignee(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	teamID := int64(1)

	oldTask := &entities.Task{
		ID:          taskID,
		Title:       "Old Title",
		Description: "Old Description",
		Status:      entities.TaskStatusTODO,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    time.Now(),
	}

	updatedTask := &entities.Task{
		ID:          taskID,
		Title:       "Old Title",
		Description: "Old Description",
		Status:      entities.TaskStatusTODO,
		AssigneeID:  3,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    oldTask.CreateAt,
	}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(oldTask, nil)

	mockRepo.EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, cases.Repository) error) error {
			return fn(ctx, mockRepo)
		})

	mockRepo.EXPECT().
		UpdateTask(ctx, gomock.Any()).
		Return(nil)

	mockRepo.EXPECT().
		AddHistoryRecord(ctx, gomock.Any()).
		Return(nil).
		AnyTimes()

	mockCache.EXPECT().
		SetTask(ctx, gomock.Any()).
		Return(nil)

	mockCache.EXPECT().
		DeleteTasksByTeam(ctx, teamID).
		Return(nil)

	err = svc.UpdateTask(ctx, userID, updatedTask)

	require.NoError(t, err)
}

func TestService_UpdateTask_UpdateDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	teamID := int64(1)
	oldTask := &entities.Task{
		ID:          taskID,
		Title:       "Old Title",
		Description: "Old Description",
		Status:      entities.TaskStatusTODO,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    time.Now(),
	}
	updatedTask := &entities.Task{
		ID:          taskID,
		Title:       "New Title",
		Description: "New Description",
		Status:      entities.TaskStatusInProgress,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    oldTask.CreateAt,
	}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(oldTask, nil)

	mockRepo.EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, cases.Repository) error) error {
			return fn(ctx, mockRepo)
		})

	mockRepo.EXPECT().
		UpdateTask(ctx, updatedTask).
		Return(errTest)

	err = svc.UpdateTask(ctx, userID, updatedTask)

	require.Error(t, err)
	require.ErrorIs(t, err, errTest)
}

func TestService_UpdateTask_AddHistoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	teamID := int64(1)
	oldTask := &entities.Task{
		ID:          taskID,
		Title:       "Old Title",
		Description: "Old Description",
		Status:      entities.TaskStatusTODO,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    time.Now(),
	}
	updatedTask := &entities.Task{
		ID:          taskID,
		Title:       "New Title",
		Description: "New Description",
		Status:      entities.TaskStatusInProgress,
		AssigneeID:  2,
		TeamID:      teamID,
		OwnerID:     userID,
		CreateAt:    oldTask.CreateAt,
	}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(oldTask, nil)

	mockRepo.EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context, cases.Repository) error) error {
			return fn(ctx, mockRepo)
		})

	mockRepo.EXPECT().
		UpdateTask(ctx, updatedTask).
		Return(nil)

	mockRepo.EXPECT().
		AddHistoryRecord(ctx, gomock.Any()).
		Return(errTest)

	err = svc.UpdateTask(ctx, userID, updatedTask)

	require.Error(t, err)
	require.ErrorIs(t, err, errTest)
}

func TestService_AddComment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	content := "Test comment"
	task := &entities.Task{ID: taskID, TeamID: 1}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(task, nil)

	mockRepo.EXPECT().
		IsMember(ctx, userID, task.TeamID).
		Return(true, nil)

	mockRepo.EXPECT().
		AddComment(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, comment *entities.TaskComment) error {
			// Проверяем, что комментарий правильный
			require.Equal(t, taskID, comment.TaskID)
			require.Equal(t, userID, comment.UserID)
			require.Equal(t, content, comment.Content)
			return nil
		})

	comment, err := svc.AddComment(ctx, userID, taskID, content)

	require.NoError(t, err)
	require.NotNil(t, comment)
	require.Equal(t, taskID, comment.TaskID)
	require.Equal(t, userID, comment.UserID)
	require.Equal(t, content, comment.Content)
}

func TestService_AddComment_UserNotInTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	content := "Test comment"
	task := &entities.Task{ID: taskID, TeamID: 1}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(task, nil)

	mockRepo.EXPECT().
		IsMember(ctx, userID, task.TeamID).
		Return(false, nil)

	comment, err := svc.AddComment(ctx, userID, taskID, content)

	require.Error(t, err)
	require.Nil(t, comment)
	require.Contains(t, err.Error(), "user is not a member of the team")
}

func TestService_AddComment_InvalidUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(0)
	taskID := int64(1)
	content := "Test comment"

	comment, err := svc.AddComment(ctx, userID, taskID, content)

	require.Error(t, err)
	require.Nil(t, comment)
	require.Contains(t, err.Error(), "userID is invalid")
}

func TestService_AddComment_InvalidTaskID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(0)
	content := "Test comment"

	comment, err := svc.AddComment(ctx, userID, taskID, content)

	require.Error(t, err)
	require.Nil(t, comment)
	require.Contains(t, err.Error(), "taskID is invalid")
}

func TestService_AddComment_EmptyContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	content := ""

	comment, err := svc.AddComment(ctx, userID, taskID, content)

	require.Error(t, err)
	require.Nil(t, comment)
	require.Contains(t, err.Error(), "content is empty")
}

func TestService_AddComment_TaskNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(999)
	content := "Test comment"

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(nil, entities.ErrNotFound)

	comment, err := svc.AddComment(ctx, userID, taskID, content)

	require.Error(t, err)
	require.Nil(t, comment)
	require.ErrorIs(t, err, entities.ErrNotFound)
}

func TestService_AddComment_AddCommentError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	taskID := int64(1)
	content := "Test comment"
	task := &entities.Task{ID: taskID, TeamID: 1}

	mockRepo.EXPECT().
		GetTaskByID(ctx, taskID).
		Return(task, nil)

	mockRepo.EXPECT().
		IsMember(ctx, userID, task.TeamID).
		Return(true, nil)

	mockRepo.EXPECT().
		AddComment(ctx, gomock.Any()).
		Return(errTest)

	comment, err := svc.AddComment(ctx, userID, taskID, content)

	require.Error(t, err)
	require.Nil(t, comment)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetCommentsByTask_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	limit := 20
	offset := 0
	expectedComments := []*entities.TaskComment{
		{TaskID: 1, UserID: 1, Content: "Comment 1"},
		{TaskID: 1, UserID: 2, Content: "Comment 2"},
	}

	mockRepo.EXPECT().
		GetCommentsByTask(ctx, taskID, limit, offset).
		Return(expectedComments, nil)

	comments, err := svc.GetCommentsByTask(ctx, taskID, limit, offset)

	require.NoError(t, err)
	require.NotNil(t, comments)
	require.Len(t, comments, 2)
	require.Equal(t, expectedComments[0].Content, comments[0].Content)
	require.Equal(t, expectedComments[1].Content, comments[1].Content)
}

func TestService_GetCommentsByTask_EmptyList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	limit := 20
	offset := 0

	mockRepo.EXPECT().
		GetCommentsByTask(ctx, taskID, limit, offset).
		Return([]*entities.TaskComment{}, nil)

	comments, err := svc.GetCommentsByTask(ctx, taskID, limit, offset)

	require.NoError(t, err)
	require.Empty(t, comments)
}

func TestService_GetCommentsByTask_InvalidTaskID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(0)
	limit := 20
	offset := 0

	comments, err := svc.GetCommentsByTask(ctx, taskID, limit, offset)

	require.Error(t, err)
	require.Nil(t, comments)
	require.Contains(t, err.Error(), "taskID is invalid")
}

func TestService_GetCommentsByTask_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	limit := 20
	offset := 0

	mockRepo.EXPECT().
		GetCommentsByTask(ctx, taskID, limit, offset).
		Return(nil, errTest)

	comments, err := svc.GetCommentsByTask(ctx, taskID, limit, offset)

	require.Error(t, err)
	require.Nil(t, comments)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetCommentsByTask_WithPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	limit := 10
	offset := 20
	expectedComments := []*entities.TaskComment{
		{TaskID: 1, UserID: 1, Content: "Comment 3"},
		{TaskID: 1, UserID: 2, Content: "Comment 4"},
	}

	mockRepo.EXPECT().
		GetCommentsByTask(ctx, taskID, limit, offset).
		Return(expectedComments, nil)

	comments, err := svc.GetCommentsByTask(ctx, taskID, limit, offset)

	require.NoError(t, err)
	require.Len(t, comments, 2)
}

func TestService_GetTaskHistory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	limit := 20
	offset := 0
	expectedHistory := []*entities.TaskHistory{
		{TaskID: 1, ChangedBy: 1, Field: "status", OldValue: strPtr("todo"), NewValue: strPtr("done")},
		{TaskID: 1, ChangedBy: 2, Field: "title", OldValue: strPtr("Old"), NewValue: strPtr("New")},
	}

	mockRepo.EXPECT().
		GetTaskHistory(ctx, taskID, limit, offset).
		Return(expectedHistory, nil)

	history, err := svc.GetTaskHistory(ctx, taskID, limit, offset)

	require.NoError(t, err)
	require.NotNil(t, history)
	require.Len(t, history, 2)
	require.Equal(t, expectedHistory[0].Field, history[0].Field)
	require.Equal(t, expectedHistory[1].Field, history[1].Field)
}

func TestService_GetTaskHistory_EmptyList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	limit := 20
	offset := 0

	mockRepo.EXPECT().
		GetTaskHistory(ctx, taskID, limit, offset).
		Return([]*entities.TaskHistory{}, nil)

	history, err := svc.GetTaskHistory(ctx, taskID, limit, offset)

	require.NoError(t, err)
	require.Empty(t, history)
}

func TestService_GetTaskHistory_InvalidTaskID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(0)
	limit := 20
	offset := 0

	history, err := svc.GetTaskHistory(ctx, taskID, limit, offset)

	require.Error(t, err)
	require.Nil(t, history)
	require.Contains(t, err.Error(), "taskID is invalid")
}

func TestService_GetTaskHistory_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	limit := 20
	offset := 0

	mockRepo.EXPECT().
		GetTaskHistory(ctx, taskID, limit, offset).
		Return(nil, errTest)

	history, err := svc.GetTaskHistory(ctx, taskID, limit, offset)

	require.Error(t, err)
	require.Nil(t, history)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTaskHistory_WithPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	taskID := int64(1)
	limit := 10
	offset := 20
	expectedHistory := []*entities.TaskHistory{
		{TaskID: 1, ChangedBy: 1, Field: "status"},
		{TaskID: 1, ChangedBy: 2, Field: "description"},
	}

	mockRepo.EXPECT().
		GetTaskHistory(ctx, taskID, limit, offset).
		Return(expectedHistory, nil)

	history, err := svc.GetTaskHistory(ctx, taskID, limit, offset)

	require.NoError(t, err)
	require.Len(t, history, 2)
}

func strPtr(s string) *string {
	return &s
}

func TestService_GetTasksByFilter_SingleTeam_FromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(1)
	filter := dto.TaskFilter{
		TeamIDs: []int64{teamID},
		Status:  nil,
		Limit:   20,
		Offset:  0,
	}
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: teamID},
		{ID: 2, Title: "Task 2", TeamID: teamID},
	}

	// Проверяем, что пользователь состоит в команде
	mockRepo.EXPECT().
		IsMember(ctx, userID, teamID).
		Return(true, nil)

	// Получаем из кеша
	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(expectedTasks, nil)

	tasks, err := svc.GetTasksByFilter(ctx, userID, filter)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
}

func TestService_GetTasksByFilter_SingleTeam_CacheMiss_DBSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(1)
	filter := dto.TaskFilter{
		TeamIDs: []int64{teamID},
		Status:  nil,
		Limit:   20,
		Offset:  0,
	}
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: teamID},
		{ID: 2, Title: "Task 2", TeamID: teamID},
	}

	mockRepo.EXPECT().
		IsMember(ctx, userID, teamID).
		Return(true, nil)

	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(nil, nil)

	mockRepo.EXPECT().
		GetTasksByTeam(ctx, teamID, 0, 0).
		Return(expectedTasks, nil)

	mockCache.EXPECT().
		SetTasksByTeam(ctx, teamID, expectedTasks).
		Return(nil)

	tasks, err := svc.GetTasksByFilter(ctx, userID, filter)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
}

func TestService_GetTasksByFilter_MultipleTeams_DB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamIDs := []int64{1, 2}
	filter := dto.TaskFilter{
		TeamIDs: teamIDs,
		Status:  nil,
		Limit:   20,
		Offset:  0,
	}
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: 1},
		{ID: 2, Title: "Task 2", TeamID: 2},
	}

	// Проверяем только первую команду (по ТЗ)
	mockRepo.EXPECT().
		IsMember(ctx, userID, teamIDs[0]).
		Return(true, nil)

	// Несколько команд → сразу в БД
	mockRepo.EXPECT().
		GetTasksByFilter(ctx, filter).
		Return(expectedTasks, nil)

	tasks, err := svc.GetTasksByFilter(ctx, userID, filter)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
}

func TestService_GetTasksByFilter_NoTeamIDs_GetUserTeams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	filter := dto.TaskFilter{
		TeamIDs: []int64{},
		Status:  nil,
		Limit:   20,
		Offset:  0,
	}
	teamIDs := []int64{1, 2}
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: 1},
		{ID: 2, Title: "Task 2", TeamID: 2},
	}

	mockRepo.EXPECT().
		GetUserTeams(ctx, userID).
		Return(teamIDs, nil)

	mockRepo.EXPECT().
		GetTasksByFilter(ctx, gomock.Any()).
		Return(expectedTasks, nil)

	tasks, err := svc.GetTasksByFilter(ctx, userID, filter)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
}

func TestService_GetTasksByFilter_NoTeamIDs_NoTeams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	filter := dto.TaskFilter{
		TeamIDs: []int64{},
		Status:  nil,
		Limit:   20,
		Offset:  0,
	}

	mockRepo.EXPECT().
		GetUserTeams(ctx, userID).
		Return([]int64{}, nil)

	tasks, err := svc.GetTasksByFilter(ctx, userID, filter)

	require.NoError(t, err)
	require.Empty(t, tasks)
}

func TestService_GetTasksByFilter_UserNotMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(999)
	filter := dto.TaskFilter{
		TeamIDs: []int64{teamID},
		Status:  nil,
		Limit:   20,
		Offset:  0,
	}

	mockRepo.EXPECT().
		IsMember(ctx, userID, teamID).
		Return(false, nil)

	tasks, err := svc.GetTasksByFilter(ctx, userID, filter)

	require.Error(t, err)
	require.Nil(t, tasks)
	require.Contains(t, err.Error(), "user is not a member of this team")
}

func TestService_GetTasksByFilter_WithStatusFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(1)
	status := "todo"
	filter := dto.TaskFilter{
		TeamIDs: []int64{teamID},
		Status:  &status,
		Limit:   20,
		Offset:  0,
	}
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: teamID, Status: "todo"},
		{ID: 2, Title: "Task 2", TeamID: teamID, Status: "todo"},
	}

	mockRepo.EXPECT().
		IsMember(ctx, userID, teamID).
		Return(true, nil)

	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(expectedTasks, nil)

	tasks, err := svc.GetTasksByFilter(ctx, userID, filter)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
}

func TestService_GetTasksByFilter_WithAssigneeFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	userID := int64(1)
	teamID := int64(1)
	assigneeID := int64(3)
	filter := dto.TaskFilter{
		TeamIDs:    []int64{teamID},
		AssigneeID: &assigneeID,
		Limit:      20,
		Offset:     0,
	}
	expectedTasks := []*entities.Task{
		{ID: 1, Title: "Task 1", TeamID: teamID, AssigneeID: 3},
		{ID: 2, Title: "Task 2", TeamID: teamID, AssigneeID: 3},
	}

	mockRepo.EXPECT().
		IsMember(ctx, userID, teamID).
		Return(true, nil)

	mockCache.EXPECT().
		GetTasksByTeam(ctx, teamID).
		Return(expectedTasks, nil)

	tasks, err := svc.GetTasksByFilter(ctx, userID, filter)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 2)
}

func TestService_GetTeamStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	expectedStats := []dto.TeamStats{
		{TeamID: 1, TeamName: "Team A", MemberCount: 5, DoneTasksLast7: 3},
		{TeamID: 2, TeamName: "Team B", MemberCount: 3, DoneTasksLast7: 1},
	}

	mockRepo.EXPECT().
		GetTeamStats(ctx).
		Return(expectedStats, nil)

	stats, err := svc.GetTeamStats(ctx)

	require.NoError(t, err)
	require.NotNil(t, stats)
	require.Len(t, stats, 2)
	require.Equal(t, expectedStats[0].TeamID, stats[0].TeamID)
	require.Equal(t, expectedStats[0].TeamName, stats[0].TeamName)
	require.Equal(t, expectedStats[0].MemberCount, stats[0].MemberCount)
	require.Equal(t, expectedStats[0].DoneTasksLast7, stats[0].DoneTasksLast7)
}

func TestService_GetTeamStats_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetTeamStats(ctx).
		Return([]dto.TeamStats{}, nil)

	stats, err := svc.GetTeamStats(ctx)

	require.NoError(t, err)
	require.Empty(t, stats)
}

func TestService_GetTeamStats_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetTeamStats(ctx).
		Return(nil, errTest)

	stats, err := svc.GetTeamStats(ctx)

	require.Error(t, err)
	require.Nil(t, stats)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetTopCreators_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	expectedCreators := []dto.TopCreator{
		{TeamID: 1, TeamName: "Team A", UserID: 10, UserName: "John", TaskCount: 15, Rank: 1},
		{TeamID: 1, TeamName: "Team A", UserID: 20, UserName: "Jane", TaskCount: 8, Rank: 2},
		{TeamID: 1, TeamName: "Team A", UserID: 30, UserName: "Bob", TaskCount: 5, Rank: 3},
		{TeamID: 2, TeamName: "Team B", UserID: 40, UserName: "Alice", TaskCount: 12, Rank: 1},
		{TeamID: 2, TeamName: "Team B", UserID: 50, UserName: "Tom", TaskCount: 7, Rank: 2},
		{TeamID: 2, TeamName: "Team B", UserID: 60, UserName: "Sue", TaskCount: 3, Rank: 3},
	}

	mockRepo.EXPECT().
		GetTopCreators(ctx).
		Return(expectedCreators, nil)

	creators, err := svc.GetTopCreators(ctx)

	require.NoError(t, err)
	require.NotNil(t, creators)
	require.Len(t, creators, 6)
	require.Equal(t, expectedCreators[0].TeamID, creators[0].TeamID)
	require.Equal(t, expectedCreators[0].TeamName, creators[0].TeamName)
	require.Equal(t, expectedCreators[0].UserID, creators[0].UserID)
	require.Equal(t, expectedCreators[0].UserName, creators[0].UserName)
	require.Equal(t, expectedCreators[0].TaskCount, creators[0].TaskCount)
	require.Equal(t, expectedCreators[0].Rank, creators[0].Rank)
}

func TestService_GetTopCreators_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetTopCreators(ctx).
		Return([]dto.TopCreator{}, nil)

	creators, err := svc.GetTopCreators(ctx)

	require.NoError(t, err)
	require.Empty(t, creators)
}

func TestService_GetTopCreators_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetTopCreators(ctx).
		Return(nil, errTest)

	creators, err := svc.GetTopCreators(ctx)

	require.Error(t, err)
	require.Nil(t, creators)
	require.ErrorIs(t, err, errTest)
}

func TestService_GetInvalidAssigneeTasks_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()
	expectedTasks := []dto.InvalidAssigneeTask{
		{
			TaskID:       1,
			TaskTitle:    "Task 1",
			TeamID:       1,
			TeamName:     "Team A",
			AssigneeID:   10,
			AssigneeName: "John",
		},
		{
			TaskID:       2,
			TaskTitle:    "Task 2",
			TeamID:       1,
			TeamName:     "Team A",
			AssigneeID:   20,
			AssigneeName: "Jane",
		},
		{
			TaskID:       3,
			TaskTitle:    "Task 3",
			TeamID:       2,
			TeamName:     "Team B",
			AssigneeID:   30,
			AssigneeName: "Bob",
		},
	}

	mockRepo.EXPECT().
		GetInvalidAssigneeTasks(ctx).
		Return(expectedTasks, nil)

	tasks, err := svc.GetInvalidAssigneeTasks(ctx)

	require.NoError(t, err)
	require.NotNil(t, tasks)
	require.Len(t, tasks, 3)
	require.Equal(t, expectedTasks[0].TaskID, tasks[0].TaskID)
	require.Equal(t, expectedTasks[0].TaskTitle, tasks[0].TaskTitle)
	require.Equal(t, expectedTasks[0].TeamID, tasks[0].TeamID)
	require.Equal(t, expectedTasks[0].TeamName, tasks[0].TeamName)
	require.Equal(t, expectedTasks[0].AssigneeID, tasks[0].AssigneeID)
	require.Equal(t, expectedTasks[0].AssigneeName, tasks[0].AssigneeName)
}

func TestService_GetInvalidAssigneeTasks_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetInvalidAssigneeTasks(ctx).
		Return([]dto.InvalidAssigneeTask{}, nil)

	tasks, err := svc.GetInvalidAssigneeTasks(ctx)

	require.NoError(t, err)
	require.Empty(t, tasks)
}

func TestService_GetInvalidAssigneeTasks_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	cfg := &mockConfig{}

	svc, err := cases.NewService(mockRepo, mockCache, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	mockRepo.EXPECT().
		GetInvalidAssigneeTasks(ctx).
		Return(nil, errTest)

	tasks, err := svc.GetInvalidAssigneeTasks(ctx)

	require.Error(t, err)
	require.Nil(t, tasks)
	require.ErrorIs(t, err, errTest)
}
