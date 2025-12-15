package repo

import (
	"database/sql"
	"fmt"
	"school-api/internal/models"
)

func FindStudentByID(id int, db *sql.DB) (*models.Student, error) {
	var s models.Student
	var className string

	err := db.QueryRow(`
        SELECT s.id, s.first_name, s.last_name, s.email, c.id AS class_id, c.name AS class_name
        FROM student s JOIN classes c ON s.class_id=c.id WHERE s.id = ?
    `, id).Scan(
		&s.ID,
		&s.FirstName,
		&s.LastName,
		&s.Email,
		&s.ClassId,
		&className,
	)

	s.Class = models.Class{
		ID:   s.ClassId,
		Name: className,
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &s, nil
}

func FindStudent(db *sql.DB, search string, filters map[string]string, sort string) ([]models.Student, error) {

	query := `
		SELECT
			s.id,
			s.first_name,
			s.last_name,
			s.email,
			c.id   AS class_id,
			c.name AS class_name
		FROM student s
		JOIN classes c ON s.class_id = c.id
		WHERE 1=1
	`

	var args []any

	// filters
	for key, val := range filters {
		if val == "" {
			continue
		}

		if key == "class" {
			query += " AND c.name = ?"
			args = append(args, val)
		} else {
			query += " AND s." + key + " = ?"
			args = append(args, val)
		}
	}

	// search
	if search != "" {
		query += `
			AND (
				s.first_name LIKE ? OR
				s.last_name  LIKE ? OR
				s.email      LIKE ? OR
				c.name       LIKE ?
			)
		`
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}

	// sorting
	if sort != "" {
		query += " ORDER BY s." + sort
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student

	for rows.Next() {
		var s models.Student
		var classID int
		var className string

		err := rows.Scan(
			&s.ID,
			&s.FirstName,
			&s.LastName,
			&s.Email,
			&classID,
			&className,
		)
		if err != nil {
			return nil, err
		}

		s.ClassId = classID
		s.Class = models.Class{
			ID:   classID,
			Name: className,
		}

		students = append(students, s)
	}

	return students, nil
}

func AddStudent(db *sql.DB, t *models.Teacher) (sql.Result, error) {

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

func UpdateStudent(db *sql.DB, existingTeacher, updateTeacher *models.Teacher, id int) (sql.Result, error) {
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

func DeleteMultipleStudents(db *sql.DB, ids []int) error {
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

func PatchMultipleStudents(db *sql.DB, ids []int, updates []map[string]any) error {

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
