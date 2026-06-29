package entities

import (
	"time"

	"github.com/pkg/errors"
)

type Team struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	OwnerID   int64     `json:"owner_id"`
	MemberIDs []int64   `json:"member_ids"`
	TaskIDs   []int64   `json:"task_ids"`
	CreateAt  time.Time `json:"create_at"`
}

func NewTeam(id int64, name string, ownerID int64, createAt time.Time) (*Team, error) {

	if name == "" {
		return nil, errors.Wrap(ErrInvalidParam, "team name is empty")
	}

	if ownerID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid team owner id")
	}

	return &Team{
		ID:        id,
		Name:      name,
		OwnerID:   ownerID,
		MemberIDs: make([]int64, 0),
		TaskIDs:   make([]int64, 0),
		CreateAt:  createAt,
	}, nil
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
	result := make([]int64, len(team.MemberIDs))
	copy(result, team.MemberIDs)
	return result
}

func (team *Team) GetTasksID() []int64 {
	result := make([]int64, len(team.TaskIDs))
	copy(result, team.TaskIDs)
	return result
}

func (team *Team) AddMemberID(memberID int64) error {
	if memberID <= 0 {
		return errors.Wrap(ErrInvalidParam, "invalid member id")
	}
	for _, id := range team.MemberIDs {
		if id == memberID {
			return nil
		}
	}
	team.MemberIDs = append(team.MemberIDs, memberID)
	return nil
}

func (team *Team) AddTaskID(taskID int64) error {
	if taskID <= 0 {
		return errors.Wrap(ErrInvalidParam, "invalid task id")
	}
	for _, id := range team.TaskIDs {
		if id == taskID {
			return nil
		}
	}

	team.TaskIDs = append(team.TaskIDs, taskID)
	return nil
}

func (team *Team) AddMemberIDs(memberIDs []int64) {
	if len(team.MemberIDs) == 0 {
		return
	}
	team.MemberIDs = append(team.MemberIDs, memberIDs...)
}

func (team *Team) AddTaskIDs(taskIDs []int64) {
	if len(team.TaskIDs) == 0 {
		return
	}
	team.TaskIDs = append(team.TaskIDs, taskIDs...)
}

func (team *Team) IsOwner(userID int64) bool {
	return team.OwnerID == userID
}

func (team *Team) IsMember(userID int64) bool {
	for _, id := range team.MemberIDs {
		if id == userID {
			return true
		}
	}
	return false
}
