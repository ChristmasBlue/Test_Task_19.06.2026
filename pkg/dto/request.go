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
	UserID int64  `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"omitempty,oneof=member admin"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title" binding:"omitempty,min=3,max=255"`
	Description string `json:"description" binding:"omitempty,min=3,max=1000"`
	Status      string `json:"status" binding:"omitempty,oneof=todo in_progress done"`
}
