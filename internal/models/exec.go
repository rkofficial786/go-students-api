package models

import "database/sql"

type Exec struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Username  string `json:"username"`

	// sensitive fields
	Password string `json:"password"`

	PasswordChangedAt    sql.NullString `json:"password_changed_at,omitempty"`
	UserCreatedAt        sql.NullString `json:"user_created_at,omitempty"`
	PasswordResetToken   sql.NullString `json:"password_reset_token"`
	PasswordTokenExpires sql.NullString `json:"password_token_expires"`

	Inactive bool   `json:"inactive"`
	Role     string `json:"role"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UpdatePasswordResponse struct {
	Token string `json:"token"`
	PasswordUpdated bool `json:"password_updated"`
}