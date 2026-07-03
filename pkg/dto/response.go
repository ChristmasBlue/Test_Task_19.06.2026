package dto

import "test_task/internal/entities"

type ErrorDTO struct {
	Status  int    `json:"status_code"`
	Message string `json:"message,omitempty"`
}

type TeamWithRole struct {
	*entities.Team
	Role string `json:"role"`
}

type TeamStats struct {
	TeamID         int64  `json:"team_id"`
	TeamName       string `json:"team_name"`
	MemberCount    int    `json:"member_count"`
	DoneTasksLast7 int    `json:"done_tasks_last_7_days"`
}

type TopCreator struct {
	TeamID    int64  `json:"team_id"`
	TeamName  string `json:"team_name"`
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	TaskCount int    `json:"task_count"`
	Rank      int    `json:"rank"`
}

type InvalidAssigneeTask struct {
	TaskID       int64  `json:"task_id"`
	TaskTitle    string `json:"task_title"`
	TeamID       int64  `json:"team_id"`
	TeamName     string `json:"team_name"`
	AssigneeID   int64  `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
}
