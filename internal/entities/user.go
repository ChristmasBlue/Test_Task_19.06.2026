package entities

import (
	"time"

	"github.com/pkg/errors"
)

type User struct {
	ID           int64
	Name         string
	Email        string
	HashPassword []byte
	TeamsID      []int64
	CreateAt     time.Time
}

func NewUser(id int64, name, email string, hashPass []byte, createAt time.Time) (*User, error) {
	newUser := &User{
		ID:           id,
		Name:         name,
		Email:        email,
		HashPassword: hashPass,
		TeamsID:      make([]int64, 0),
		CreateAt:     createAt,
	}

	if newUser.Name == "" {
		return nil, errors.Wrap(ErrInvalidParam, "name is empty")
	}

	if newUser.Email == "" {
		return nil, errors.Wrap(ErrInvalidParam, "email is empty")
	}

	if len(newUser.HashPassword) == 0 {
		return nil, errors.Wrap(ErrInvalidParam, "hash password is empty")
	}

	return newUser, nil
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
	result := make([]int64, len(user.TeamsID))
	copy(result, user.TeamsID)
	return result
}

func (user *User) AddTeamID(teamID int64) error {
	if teamID <= 0 {
		return errors.Wrap(ErrInvalidParam, "invalid team id")
	}
	for _, id := range user.TeamsID {
		if id == teamID {
			return nil
		}
	}
	user.TeamsID = append(user.TeamsID, teamID)
	return nil
}
