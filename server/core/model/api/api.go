package model_api

type Testcase struct {
	ID        string
	SessionID string
	Name      string
	Status    string
	CreatedAt string
}

type Session struct {
	ID          string
	Description string
	Status      string
	CreatedAt   string
}

type Group struct {
	Status string
	Group  string
}
