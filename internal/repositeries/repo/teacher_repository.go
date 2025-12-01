package repo

import (
	"database/sql"
	"school-api/internal/models"
)

func FindTeacherByID(id int, db *sql.DB) (*models.Teacher, error) {
	var t models.Teacher

	err := db.QueryRow(`
        SELECT id, first_name, last_name, email, class, subject 
        FROM teachers WHERE id = ?
    `, id).Scan(
		&t.ID,
		&t.FirstName,
		&t.LastName,
		&t.Email,
		&t.Class,
		&t.Subject,
	)

	if err == sql.ErrNoRows {
		return nil, nil // not found
	}

	if err != nil {
		return nil, err
	}

	return &t, nil
}
