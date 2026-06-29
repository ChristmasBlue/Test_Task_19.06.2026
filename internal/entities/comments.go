package entities

import (
	"time"

	"github.com/pkg/errors"
)

type TaskComment struct {
	//ID        int64     `json:"id"`
	TaskID   int64     `json:"task_id"`
	UserID   int64     `json:"user_id"`
	Content  string    `json:"content"`
	CreateAt time.Time `json:"create_at"`
}

func NewTaskComment(taskID, userID int64, content string, createAt time.Time) (*TaskComment, error) {
	if taskID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "taskID is invalid")
	}
	if userID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "userID is invalid")
	}
	if content == "" {
		return nil, errors.Wrap(ErrInvalidParam, "content is empty")
	}

	return &TaskComment{
		//		ID:        id,
		TaskID:   taskID,
		UserID:   userID,
		Content:  content,
		CreateAt: createAt,
	}, nil
}
