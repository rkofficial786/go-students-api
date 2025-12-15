package repo

import (
	"database/sql"
	"fmt"
	"school-api/internal/models"
)

func FindTeacherByID(id int, db *sql.DB) (*models.Teacher, error) {
	var t models.Teacher

	err := db.QueryRow(`
        SELECT t.id, t.first_name, t.last_name, t.email, c.name AS class, t.subject 
        FROM teachers t JOIN classes c ON t.class_id=c.id WHERE t.id = ?
    `, id).Scan(
		&t.ID,
		&t.FirstName,
		&t.LastName,
		&t.Email,
		&t.Class,
		&t.Subject,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func FindTeacher(db *sql.DB, search string, filters map[string]string, sort string) ([]models.Teacher, error) {

	query := `
		SELECT 
			t.id, 
			t.first_name, 
			t.last_name, 
			t.email, 
			c.name AS class,
			t.subject
		FROM teachers t
		JOIN classes c ON t.class_id = c.id
		WHERE 1=1
	`

	var args []any

	for key, val := range filters {
		if val == "" {
			continue
		}

		if key == "class" {
			query += " AND c.name = ?"
			args = append(args, val)
		} else {
			query += " AND t." + key + " = ?"
			args = append(args, val)
		}
	}

	if search != "" {
		query += `
			AND (
				t.first_name LIKE ? OR
				t.last_name LIKE ? OR
				t.subject LIKE ? OR
				t.email LIKE ? OR
				c.name LIKE ?
			)
		`
		searchPattern := "%" + search + "%"
		args = append(args,
			searchPattern,
			searchPattern,
			searchPattern,
			searchPattern,
			searchPattern,
		)
	}

	if sort != "" {
		query += " ORDER BY t." + sort
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teachers []models.Teacher
	for rows.Next() {
		var t models.Teacher
		err := rows.Scan(
			&t.ID,
			&t.FirstName,
			&t.LastName,
			&t.Email,
			&t.Class, // this is c.name
			&t.Subject,
		)
		if err != nil {
			return nil, err
		}
		teachers = append(teachers, t)
	}

	return teachers, nil
}

func AddTeacher(db *sql.DB, t *models.Teacher) (sql.Result, error) {

	stmt, err := db.Prepare("INSERT INTO teachers (first_name,last_name,email,class,subject) VALUES (?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(t.FirstName, t.LastName, t.Email, t.Class, t.Subject)
	if err != nil {
		return nil, err
	}

	return res, nil

}

func UpdateTeacher(db *sql.DB, existingTeacher, updateTeacher *models.Teacher, id int) (sql.Result, error) {
	err := db.QueryRow("SELECT id,first_name,last_name,email,subject,class FROM teachers WHERE id= ?", id).
		Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Subject, &existingTeacher.Class)

	if err != nil {

		return nil, err
	}

	updateTeacher.ID = existingTeacher.ID
	// Simple conditional updates
	if updateTeacher.FirstName == "" {
		updateTeacher.FirstName = existingTeacher.FirstName
	}
	if updateTeacher.LastName == "" {
		updateTeacher.LastName = existingTeacher.LastName
	}
	if updateTeacher.Email == "" {
		updateTeacher.Email = existingTeacher.Email
	}
	if updateTeacher.Subject == "" {
		updateTeacher.Subject = existingTeacher.Subject
	}
	if updateTeacher.Class == "" {
		updateTeacher.Class = existingTeacher.Class
	}

	_, err = db.Exec("UPDATE teachers SET first_name=?, last_name=?, email=?, subject=?, class=? WHERE id=?",
		updateTeacher.FirstName, updateTeacher.LastName, updateTeacher.Email, updateTeacher.Subject, updateTeacher.Class, id)

	if err != nil {

		return nil, err
	}

	return nil, nil

}

func DeleteMultipleTeachers(db *sql.DB, ids []int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	for _, id := range ids {
		result, err := tx.Exec("DELETE FROM teachers WHERE id = ?", id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete ID %d: %w", id, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}

		if rowsAffected == 0 {
			tx.Rollback()
			return fmt.Errorf("teacher with ID %d not found", id)
		}
	}

	return tx.Commit()
}

func PatchMultipleTeachers(db *sql.DB, ids []int, updates []map[string]any) error {

	tx, err := db.Begin()
	if err != nil {

		return nil
	}

	for _, update := range updates {

		var id int
		if idFloat, ok := update["id"].(float64); ok {
			id = int(idFloat)
		} else if idInt, ok := update["id"].(int); ok {
			id = idInt
		} else {
			tx.Rollback()

			return err
		}

		var teacherFromDB models.Teacher
		// Note: Using 'tx' for QueryRow ensures we are inside the transaction
		err := tx.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).
			Scan(&teacherFromDB.ID, &teacherFromDB.FirstName, &teacherFromDB.LastName, &teacherFromDB.Email, &teacherFromDB.Class, &teacherFromDB.Subject)

		if err == sql.ErrNoRows {
			tx.Rollback()

			return fmt.Errorf("teacher with id %d not found, %w", id, err)
		} else if err != nil {
			tx.Rollback()
			return err
		}

		if firstName, ok := update["first_name"].(string); ok {
			teacherFromDB.FirstName = firstName
		}
		if lastName, ok := update["last_name"].(string); ok {
			teacherFromDB.LastName = lastName
		}
		if email, ok := update["email"].(string); ok {
			teacherFromDB.Email = email
		}
		if class, ok := update["class"].(string); ok {
			teacherFromDB.Class = class
		}
		if subject, ok := update["subject"].(string); ok {
			teacherFromDB.Subject = subject
		}

		_, err = tx.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ?, class = ?, subject = ? WHERE id = ?",
			teacherFromDB.FirstName, teacherFromDB.LastName, teacherFromDB.Email, teacherFromDB.Class, teacherFromDB.Subject, id)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
