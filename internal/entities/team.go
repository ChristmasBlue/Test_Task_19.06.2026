package entities

import (
	"time"

	"github.com/pkg/errors"
)

type Team struct {
	ID        int64
	Name      string
	OwnerID   int64
	MembersID []int64
	TasksID   []int64
	CreatedAt time.Time
}

func NewTeam(id int64, name string, ownerID int64, createAt time.Time) (*Team, error) {
	newTeam := &Team{
		ID:        id,
		Name:      name,
		OwnerID:   ownerID,
		MembersID: make([]int64, 0),
		TasksID:   make([]int64, 0),
		CreatedAt: createAt,
	}

	if newTeam.ID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid team id")
	}

	if newTeam.Name == "" {
		return nil, errors.Wrap(ErrInvalidParam, "team name is empty")
	}

	if newTeam.OwnerID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid team owner id")
	}

	return newTeam, nil
}

func (team *Team) GetID() int64 {
	return team.ID
}

func (team *Team) GetName() string {
	return team.Name
}

func (team *Team) GetOwnerID() int64 {
	return team.OwnerID
}

func (team *Team) GetMembersID() []int64 {
	result := make([]int64, len(team.MembersID))
	copy(result, team.MembersID)
	return result
}

func (team *Team) GetTasksID() []int64 {
	result := make([]int64, len(team.TasksID))
	copy(result, team.TasksID)
	return result
}

func (team *Team) AddMemberID(memberID int64) error {
	if memberID <= 0 {
		return errors.Wrap(ErrInvalidParam, "invalid member id")
	}
	for _, id := range team.MembersID {
		if id == memberID {
			return nil
		}
	}
	team.MembersID = append(team.MembersID, memberID)
	return nil
}

func (team *Team) AddTaskID(taskID int64) error {
	if taskID <= 0 {
		return errors.Wrap(ErrInvalidParam, "invalid task id")
	}
	for _, id := range team.TasksID {
		if id == taskID {
			return nil
		}
	}

	team.TasksID = append(team.TasksID, taskID)
	return nil
}
