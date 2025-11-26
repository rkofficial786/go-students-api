package models

type Student struct {
	ID int `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName string `json:"last_name,omitempty"`
	Age  int    `json:"age,omitempty"`
	Email string `json:"email,omitempty"`
	Class string `json:"class,omitempty"`
}