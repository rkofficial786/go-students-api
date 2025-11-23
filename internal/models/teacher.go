package models

type Teacher struct {
	ID        int `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Subject   string `json:"subject"`
	Email     string `json:"email"`
	Class     string `json:"class"`
}
