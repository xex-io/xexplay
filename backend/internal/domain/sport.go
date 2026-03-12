package domain

type Sport struct {
	Key      string `json:"key"`
	Group    string `json:"group"`
	Title    string `json:"title"`
	IsActive bool   `json:"is_active"`
}
