package entities

type Filter struct {
	TeamID     *int64
	AssigneeID *int64
	Status     *string
}

func NewFilter(teamID, assigneeID int64, status string) *Filter {
	newFilter := &Filter{}

	if teamID > 0 {
		newFilter.TeamID = &teamID
	}

	if assigneeID > 0 {
		newFilter.AssigneeID = &assigneeID
	}

	if status != "" {
		newFilter.Status = &status
	}
	return newFilter
}
