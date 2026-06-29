package entities

import (
	"time"

	"github.com/pkg/errors"
)

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	HashPassword []byte    `json:"-"`
	TeamIDs      []int64   `json:"team_ids"`
	CreateAt     time.Time `json:"create_at"`
}

func NewUser(id int64, name, email string, hashPass []byte, createAt time.Time) (*User, error) {

	if name == "" {
		return nil, errors.Wrap(ErrInvalidParam, "name is empty")
	}

	if email == "" {
		return nil, errors.Wrap(ErrInvalidParam, "email is empty")
	}

	if len(hashPass) == 0 {
		return nil, errors.Wrap(ErrInvalidParam, "hash password is empty")
	}

	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		HashPassword: hashPass,
		TeamIDs:      make([]int64, 0),
		CreateAt:     createAt,
	}, nil
}

func (user *User) GetID() int64 {
	return user.ID
}

func (user *User) GetEmail() string {
	return user.Email
}

func (user *User) GetHashPass() []byte {
	result := make([]byte, len(user.HashPassword))
	copy(result, user.HashPassword)

	return result
}

func (user *User) GetTeamsID() []int64 {
	result := make([]int64, len(user.TeamIDs))
	copy(result, user.TeamIDs)
	return result
}

func (user *User) AddTeamID(teamID int64) error {
	if teamID <= 0 {
		return errors.Wrap(ErrInvalidParam, "invalid team id")
	}
	for _, id := range user.TeamIDs {
		if id == teamID {
			return nil
		}
	}
	user.TeamIDs = append(user.TeamIDs, teamID)
	return nil
}

func (user *User) AddTeamIDs(teamID []int64) {
	if len(teamID) == 0 {
		return
	}
	user.TeamIDs = append(user.TeamIDs, teamID...)
}

func (user *User) IsMember(teamID int64) bool {
	for _, id := range user.TeamIDs {
		if id == teamID {
			return true
		}
	}
	return false
}
