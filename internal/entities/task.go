package entities

import (
	"time"

	"github.com/pkg/errors"
)

type Task struct {
	ID          int64
	AssigneeID  int64
	OwnerID     int64
	TeamID      int64
	Title       string
	Description string
	Status      string
	Comments    []*Comment
	CreateAt    time.Time
}

type Comment struct {
	UserID    int64
	Comm      string
	CreatedAt time.Time
}

func NewTask(taskID, ownerID, teamID, assigneeID int64, title, description, status string, createAt time.Time) (*Task, error) {
	newTask := &Task{
		ID:          taskID,
		Title:       title,
		Description: description,
		Status:      status,
		AssigneeID:  assigneeID,
		OwnerID:     ownerID,
		TeamID:      teamID,
		Comments:    make([]*Comment, 0),
		CreateAt:    createAt,
	}

	if newTask.ID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid task id")
	}

	if newTask.Title == "" {
		return nil, errors.Wrap(ErrInvalidParam, "task title is empty")
	}

	if newTask.OwnerID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid task owner id")
	}

	if newTask.AssigneeID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid task assignee id")
	}

	return newTask, nil
}

func NewComment(userID int64, comm string) (*Comment, error) {
	if userID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid user id")
	}

	if comm == "" {
		return nil, errors.Wrap(ErrInvalidParam, "comment is empty")
	}

	newComment := &Comment{
		UserID: userID,
		Comm:   comm,
	}

	return newComment, nil
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

func (task *Task) GetComments() []*Comment {
	return task.Comments
}

func (task *Task) AddComment(comm *Comment) error {
	task.Comments = append(task.Comments, comm)
	return nil
}

func (task *Task) GetTeamID() int64 {
	return task.TeamID
}

func (c *Comment) SetCreateAt(createAt time.Time) {
	c.CreatedAt = createAt
}

func (task *Task) SetStatus(status string) error {
	if status == StatusCreated || status == StatusClosed || status == StatusInProgress || status == StatusCompleted || status == StatusTODO {
		task.Status = status
		return nil
	}
	return errors.Wrap(ErrInvalidParam, "invalid task status")
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
