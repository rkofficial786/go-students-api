package repo

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"fmt"
	"school-api/internal/models"
	"strings"

	"golang.org/x/crypto/argon2"
)

func FindExecByID(id int, db *sql.DB) (*models.Exec, error) {
	var e models.Exec

	err := db.QueryRow(`
        SELECT
			id,
			first_name,
			last_name,
			email,
			username,
			inactive ,
			role,
			user_created_at ,
			password_changed_at
			
		FROM execs WHERE id = ?
    `, id).Scan(
		&e.ID,
		&e.FirstName,
		&e.LastName,
		&e.Email,
		&e.Username,
		&e.Inactive,
		&e.Role,
		&e.UserCreatedAt,
		&e.PasswordChangedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &e, nil
}

func FindExec(db *sql.DB, search string, filters map[string]string, sort string) ([]models.Exec, error) {

	query := `
		SELECT
			id,
			first_name,
			last_name,
			email,
			username,
			inactive ,
			role,
			user_created_at ,
			password_changed_at
			
		FROM execs 
		WHERE 1=1
	`

	var args []any

	// filters
	for key, val := range filters {
		if val == "" {
			continue
		}

		query += " AND " + key + " = ?"
		args = append(args, val)

	}

	// search
	if search != "" {
		query += `
			AND (
				first_name LIKE ? OR
				last_name  LIKE ? OR
				email      LIKE ? OR
				username      LIKE ? 
				
			)
		`
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern, pattern)
	}

	// sorting
	if sort != "" {
		query += " ORDER BY " + sort
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var execs []models.Exec

	for rows.Next() {
		var e models.Exec

		err := rows.Scan(
			&e.ID,
			&e.Email,
			&e.LastName,
			&e.FirstName,
			&e.Username,
			&e.Inactive,
			&e.Role,
			&e.UserCreatedAt,
			&e.PasswordChangedAt,
		)
		if err != nil {
			return nil, err
		}

		execs = append(execs, e)
	}

	return execs, nil
}

func AddExec(db *sql.DB, t *models.Exec) (sql.Result, error) {

	stmt, err := db.Prepare("INSERT INTO execs (first_name,last_name,email,username,password,role,inactive) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(t.FirstName, t.LastName, t.Email, t.Username, t.Password, t.Role, t.Inactive)
	if err != nil {
		return nil, err
	}

	return res, nil

}

func UpdateExec(db *sql.DB, existingStudent, updateStudent *models.Student, id int) (sql.Result, error) {
	err := db.QueryRow("SELECT id,first_name,last_name,email,class_id FROM student WHERE id= ?", id).
		Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.ClassId)

	if err != nil {
		return nil, err
	}

	updateStudent.ID = existingStudent.ID
	// Simple conditional updates
	if updateStudent.FirstName == "" {
		updateStudent.FirstName = existingStudent.FirstName
	}
	if updateStudent.LastName == "" {
		updateStudent.LastName = existingStudent.LastName
	}
	if updateStudent.Email == "" {
		updateStudent.Email = existingStudent.Email
	}

	if updateStudent.ClassId == 0 {
		updateStudent.ClassId = existingStudent.ClassId
	}

	_, err = db.Exec("UPDATE student SET first_name=?, last_name=?, email=?, class_id=? WHERE id=?",
		updateStudent.FirstName, updateStudent.LastName, updateStudent.Email, updateStudent.ClassId, id)

	if err != nil {

		return nil, err
	}

	return nil, nil

}

func CheckExecExists(db *sql.DB, email string, id int) (bool, error) {
	var tmp int
	var err error
	if email != "" {
		err = db.QueryRow("SELECT id FROM student WHERE email=?", email).Scan(&tmp)
	} else if id != 0 {
		err = db.QueryRow("SELECT id FROM student WHERE id=?", id).Scan(&tmp)
	} else {
		return false, err
	}

	if err == sql.ErrNoRows {
		return false, err
	}

	return true, err

}

func VerifyPassword(currentPass string, userPass string) (bool, error) {
	parts := strings.Split(userPass, ".")

	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)

	if err != nil {

		return false, err
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)

	if err != nil {

		return false, err
	}

	hash := argon2.IDKey([]byte(currentPass), salt, 1, 64*1024, 4, 32)

	if len(hash) != len(hashedPassword) {
		return false, err
	}

	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {

	} else {
		return false, err
	}
	return true, nil
}

func EncryptPassword(password string) (string, error) {

	salt := make([]byte, 16)
	_, err := rand.Read(salt)

	if err != nil {

		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)

	return encodedHash, nil

}
