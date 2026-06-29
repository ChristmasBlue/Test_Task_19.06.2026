package entities

import "github.com/pkg/errors"

type TeamMember struct {
	UserID int64  `json:"user_id"`
	TeamID int64  `json:"team_id"`
	Role   string `json:"role"`
}

func NewTeamMember(userID int64, teamID int64, role string) (*TeamMember, error) {
	if userID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "userID is invalid")
	}

	if teamID <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "teamID is invalid")
	}

	if role == "" {
		role = "member"
	}

	if role != "member" && role != "admin" && role != "owner" {
		return nil, errors.Wrap(ErrInvalidParam, "role is invalid")
	}

	return &TeamMember{
		UserID: userID,
		TeamID: teamID,
		Role:   role,
	}, nil
}

func (t *TeamMember) GetRole() string {
	return t.Role
}
