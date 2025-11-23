package domain

type Team struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}
