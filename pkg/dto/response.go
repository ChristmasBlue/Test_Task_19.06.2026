package dto

type ErrorDTO struct {
	Status  int    `json:"status_code"`
	Message string `json:"message,omitempty"`
}
