package entities

import (
	"time"

	"github.com/pkg/errors"
)

type Task struct {
	ID          int64      `json:"id"`
	AssigneeID  int64      `json:"assignee_id"`
	OwnerID     int64      `json:"owner_id"`
	TeamID      int64      `json:"team_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreateAt    time.Time  `json:"create_at"`
	CompleteAt  *time.Time `json:"complete_at"`
}

var (
	TaskStatusInProgress = "in_progress"
	TaskStatusDone       = "done"
	TaskStatusTODO       = "todo"
)

func NewTask(taskID, ownerID, teamID, assigneeID int64, title, description, status string, createAt time.Time) (*Task, error) {

	if title == "" {
		return nil, errors.Wrap(ErrInvalidParam, "task title is empty")
	}

	if ownerID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid task owner id")
	}

	if assigneeID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid task assignee id")
	}

	return &Task{
		ID:          taskID,
		Title:       title,
		Description: description,
		Status:      status,
		AssigneeID:  assigneeID,
		OwnerID:     ownerID,
		TeamID:      teamID,
		CreateAt:    createAt,
	}, nil
}

func (task *Task) GetID() int64 {
	return task.ID
}

func (task *Task) GetTitle() string {
	return task.Title
}

func (task *Task) GetDescription() string {
	return task.Description
}

func (task *Task) GetStatus() string {
	return task.Status
}

func (task *Task) GetAssigneeID() int64 {
	return task.AssigneeID
}

func (task *Task) GetOwnerID() int64 {
	return task.OwnerID
}

func (task *Task) GetTeamID() int64 {
	return task.TeamID
}

func (task *Task) SetStatus(status string) error {
	if !isValidTaskStatus(status) {
		return errors.Wrap(ErrInvalidParam, "invalid task status")
	}

	if status == TaskStatusDone && task.Status != TaskStatusDone {
		now := time.Now()
		task.CompleteAt = &now
	}

	// ✅ Если статус меняется с "done" на что-то другое — очищаем completed_at
	if task.Status == TaskStatusDone && status != TaskStatusDone {
		task.CompleteAt = nil
	}

	task.Status = status
	return nil
}

func (task *Task) SetTitle(title string) {
	task.Title = title
}

func (task *Task) SetDescription(description string) {
	task.Description = description
}

func (task *Task) SetAssigneeID(assigneeID int64) {
	task.AssigneeID = assigneeID
}

func (task *Task) SetID(id int64) {
	task.ID = id
}

func (task *Task) SetCompeteAt(completeAt *time.Time) {
	task.CompleteAt = completeAt
}

func isValidTaskStatus(status string) bool {
	switch status {
	case TaskStatusTODO, TaskStatusInProgress, TaskStatusDone:
		return true
	default:
		return false
	}
}
