package entities

import (
	"time"

	"github.com/pkg/errors"
)

type TaskHistory struct {
	//ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	ChangedBy int64     `json:"changed_by"`
	Field     string    `json:"field"`
	OldValue  *string   `json:"old_value,omitempty"`
	NewValue  *string   `json:"new_value,omitempty"`
	ChangeAt  time.Time `json:"change_at"`
}

func NewTaskHistory(taskID, changedBy int64, field string, oldValue, newValue *string, changeAt time.Time) (*TaskHistory, error) {
	if taskID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "taskID is invalid")
	}
	if changedBy <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "changedBy is invalid")
	}
	if field == "" {
		return nil, errors.Wrap(ErrInvalidParam, "field is invalid")
	}

	return &TaskHistory{
		TaskID:    taskID,
		ChangedBy: changedBy,
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		ChangeAt:  changeAt,
	}, nil
}
