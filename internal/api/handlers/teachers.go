package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"school-api/internal/models"
	"school-api/internal/repositeries/sqlconnect"
	"strconv"

	"strings"

	"net/http"
)

func TeachersHandlers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("teachersHandlers called with method:", r.Method)

	switch r.Method {
	case http.MethodGet:
		fmt.Println("GET /teachers/")
		getTeachersHandler(w, r)
	case http.MethodPost:
		fmt.Println("POST /teachers/")
		addTeacherHandler(w, r)
	case http.MethodDelete:
		fmt.Fprintln(w, "Hello, World! Delete Teachers")
	case http.MethodPut:
		fmt.Fprintln(w, "Hello, World! Put Teachers")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	path = strings.Trim(path, "/")

	if path == "" {

		search := r.URL.Query().Get("search")

		params := map[string]string{
			"class":      "class",
			"first_name": "first_name",
			"last_name":  "last_name",
			"subject":    "subject",
			"email":      "email",
		}

		query := "SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE 1=1"
		var args []any

		for param, dbField := range params {
			value := r.URL.Query().Get(param)
			if value != "" {
				query += " AND " + dbField + " = ?"
				args = append(args, value)
			}
		}

		if search != "" {
			query += " AND (first_name LIKE ? OR last_name LIKE ? OR subject LIKE ? OR email LIKE ? OR class LIKE ? )"
			searchPattern := "%" + search + "%"
			args = append(args, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
		}

		rows, err := db.Query(query, args...)

		if err != nil {
			http.Error(w, "Database query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var teachers []models.Teacher

		for rows.Next() {
			var t models.Teacher

			err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Email, &t.Class, &t.Subject)
			if err != nil {
				http.Error(w, "Error scanning row", http.StatusInternalServerError)
				return
			}

			teachers = append(teachers, t)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"count":   len(teachers),
			"message": "teachers fetched successfully",
			"data":    teachers,
		})
		return

	}

	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var t models.Teacher

	err = db.QueryRow("SELECT id, first_name ,last_name, email,class,subject FROM teachers WHERE id= ?", id).Scan(&t.ID, &t.FirstName, &t.LastName, &t.Email, &t.Class, &t.Subject)

	if err == sql.ErrNoRows {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return

	} else if err != nil {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "teacher fetched successfully",
		"data":    t,
	})

}

func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	defer db.Close()

	var teacher models.Teacher
	_ = json.NewDecoder(r.Body).Decode(&teacher)

	fmt.Println(teacher, "models.Teacher")

	stmt, err := db.Prepare("INSERT INTO teachers (first_name,last_name,email,class,subject) VALUES (?,?,?,?,?)")

	if err != nil {
		fmt.Println("SQL Prepare Error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	defer stmt.Close()

	res, err := stmt.Exec(teacher.FirstName, teacher.LastName, teacher.Email, teacher.Class, teacher.Subject)
	if err != nil {
		http.Error(w, "Failed to add teacher", http.StatusInternalServerError)
		return
	}
	lastId, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve last insert id", http.StatusInternalServerError)
		return
	}

	teacher.ID = int(lastId)

	response := struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    models.Teacher `json:"data"`
	}{
		Status:  "success",
		Message: "teacher added successfully",
		Data:    teacher,
	}
	json.NewEncoder(w).Encode(response)

}
