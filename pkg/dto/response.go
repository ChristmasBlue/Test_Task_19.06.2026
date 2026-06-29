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
