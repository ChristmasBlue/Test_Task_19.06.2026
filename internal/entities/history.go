package entities

import (
	"time"

	"github.com/pkg/errors"
)

type History struct {
	Records []*Record
}

type Record struct {
	UserID      int64
	TaskID      int64
	Status      string
	Title       string
	Description string
	CreatedAt   time.Time
}

func NewHistory() *History {
	return &History{
		Records: make([]*Record, 0),
	}
}

func NewRecord(userID int64, taskID int64, status, title, description string, createAt time.Time) (*Record, error) {
	if userID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid user id")
	}

	if taskID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "invalid task id")
	}

	return &Record{
		UserID:      userID,
		TaskID:      taskID,
		Status:      status,
		Title:       title,
		Description: description,
		CreatedAt:   createAt,
	}, nil
}

func (h *History) GetRecords() []*Record {
	return h.Records
}

func (h *History) AddRecord(record *Record) {
	h.Records = append(h.Records, record)
}

func (r *Record) SetCreateAt(createAt time.Time) {
	r.CreatedAt = createAt
}

func (r *Record) SetUserID(userID int64) {
	r.UserID = userID
}

func (r *Record) SetTaskID(taskID int64) {
	r.TaskID = taskID
}

func (r *Record) SetStatus(status string) {
	r.Status = status
}

func (r *Record) SetTitle(title string) {
	r.Title = title
}

func (r *Record) SetDescription(description string) {
	r.Description = description
}

func (r *Record) GetUserID() int64 {
	return r.UserID
}
func (r *Record) GetTaskID() int64 {
	return r.TaskID
}
func (r *Record) GetStatus() string {
	return r.Status
}
func (r *Record) GetTitle() string {
	return r.Title
}
func (r *Record) GetDescription() string {
	return r.Description
}
func (r *Record) GetCreatedAt() time.Time {
	return r.CreatedAt
}
