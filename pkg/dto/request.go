package dto

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CommentRequest struct {
	Comment string `json:"comment"`
}

type TaskRequest struct {
	Title       string `json:"title" form:"title"`
	Description string `json:"description" form:"description"`
	TeamID      int64  `json:"team_id" form:"team_id"`
	AssigneeID  int64  `json:"assignee_id" form:"assignee_id"`
}

type TeamRequest struct {
	InviteUserID int64  `json:"invite_user_id"`
	UserID       int64  `json:"user_id"`
	Name         string `json:"team_name"`
}

type TeamInviteRequest struct {
	InviteUserID int64 `json:"invite_user_id"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	TeamID      int64  `json:"team_id" binding:"required"`
	AssigneeID  int64  `json:"assignee_id" binding:"required"`
	Status      string `json:"status" binding:"required"`
}
