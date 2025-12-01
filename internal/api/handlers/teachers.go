package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"school-api/internal/models"
	"school-api/internal/repositeries/db"
	"school-api/internal/repositeries/repo"
	"school-api/internal/repositeries/sqlconnect"
	"school-api/pkg/utils"
	"strconv"
	"strings"
)

// --- Helper Functions ---

func isValidSortField(field string) bool {
	validFields := map[string]bool{
		"first_name": true, "last_name": true, "email": true, "class": true, "subject": true,
	}
	return validFields[field]
}

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func addSort(r *http.Request, query string) string {
	sortParams := r.URL.Query()["sortBy"]
	if len(sortParams) > 0 {
		query += " ORDER BY"
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if !isValidSortField(field) || !isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				query += ","
			}
			query += " " + field + " " + order
		}
	}
	return query
}

// --- Handlers ---

func GetTeacherByIdHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.WriteJSONError(w, err, "Error Connecting db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteJSONError(w, err, "Invalid teacher ID format", http.StatusBadRequest)
		return
	}

	teacher, err := repo.FindTeacherByID(id, db.DB)

	if teacher == nil {
		utils.WriteJSONError(w, nil, "Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		utils.WriteJSONError(w, err, "Error fetching teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Teacher fetched successfully",
		"data":    teacher,
	})
}

func GetTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.WriteJSONError(w, err, "Error Connecting db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

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

	query = addSort(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		utils.WriteJSONError(w, err, "Failed to execute query", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var teachers []models.Teacher
	for rows.Next() {
		var t models.Teacher
		err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Email, &t.Class, &t.Subject)
		if err != nil {
			utils.WriteJSONError(w, err, "Error reading row data", http.StatusInternalServerError)
			return
		}
		teachers = append(teachers, t)
	}

	// Return empty array instead of null if no teachers found
	if teachers == nil {
		teachers = []models.Teacher{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"count":   len(teachers),
		"message": "Teachers fetched successfully",
		"data":    teachers,
	})
}

func AddTeacherHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.WriteJSONError(w, err, "Error Connecting db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var teacher models.Teacher

	if err := json.NewDecoder(r.Body).Decode(&teacher); err != nil {
		utils.WriteJSONError(w, err, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO teachers (first_name,last_name,email,class,subject) VALUES (?,?,?,?,?)")
	if err != nil {
		utils.WriteJSONError(w, err, "Database preparation error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(teacher.FirstName, teacher.LastName, teacher.Email, teacher.Class, teacher.Subject)
	if err != nil {
		utils.WriteJSONError(w, err, "Failed to insert teacher (Duplicate email likely)", http.StatusInternalServerError)
		return
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		utils.WriteJSONError(w, err, "Failed to retrieve ID", http.StatusInternalServerError)
		return
	}

	teacher.ID = int(lastId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status  string         `json:"status"`
		Message string         `json:"message"`
		Data    models.Teacher `json:"data"`
	}{
		Status:  "success",
		Message: "Teacher added successfully",
		Data:    teacher,
	}
	json.NewEncoder(w).Encode(response)
}

func UpdateTeacherHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteJSONError(w, err, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	var updateTeacher models.Teacher
	if err := json.NewDecoder(r.Body).Decode(&updateTeacher); err != nil {
		utils.WriteJSONError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	db, err := db.New()
	if err != nil {
		utils.WriteJSONError(w, err, "Error Connecting db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher
	err = db.QueryRow("SELECT id,first_name,last_name,email,subject,class FROM teachers WHERE id= ?", id).
		Scan(&existingTeacher.ID, &existingTeacher.FirstName, &existingTeacher.LastName, &existingTeacher.Email, &existingTeacher.Subject, &existingTeacher.Class)

	if err == sql.ErrNoRows {
		utils.WriteJSONError(w, nil, "Teacher not found", http.StatusNotFound)
		return
	} else if err != nil {
		utils.WriteJSONError(w, err, "Database lookup error", http.StatusInternalServerError)
		return
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
		utils.WriteJSONError(w, err, "Failed to update teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Teacher updated successfully",
		"data":    updateTeacher,
	})
}

func DeleteTeacherHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteJSONError(w, err, "Invalid teacher ID", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		utils.WriteJSONError(w, err, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		utils.WriteJSONError(w, err, "Failed to delete teacher", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.WriteJSONError(w, err, "Error checking deletion status", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		utils.WriteJSONError(w, nil, "Teacher not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteMupltipleTeachersHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		utils.WriteJSONError(w, err, "Invalid JSON body (expecting array of IDs)", http.StatusBadRequest)
		return
	}

	db, err := db.New()
	if err != nil {
		utils.WriteJSONError(w, err, "Error Connecting db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		utils.WriteJSONError(w, err, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}

	for _, id := range ids {
		result, err := tx.Exec("DELETE FROM teachers WHERE id = ? ", id)
		if err != nil {
			tx.Rollback()
			utils.WriteJSONError(w, err, fmt.Sprintf("Failed to delete ID %d", id), http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			utils.WriteJSONError(w, err, "Error checking rows affected", http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			tx.Rollback()
			utils.WriteJSONError(w, nil, fmt.Sprintf("Teacher with ID %d not found", id), http.StatusNotFound)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.WriteJSONError(w, err, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func PatchMultipleTeachersHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.WriteJSONError(w, err, "Error Connecting db", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var updates []map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.WriteJSONError(w, err, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		utils.WriteJSONError(w, err, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}

	for _, update := range updates {
		// FIX: JSON numbers often decode as float64 in interface{}.
		// We handle both float64 and int to be safe.
		var id int
		if idFloat, ok := update["id"].(float64); ok {
			id = int(idFloat)
		} else if idInt, ok := update["id"].(int); ok {
			id = idInt
		} else {
			tx.Rollback()
			utils.WriteJSONError(w, nil, "Invalid or missing teacher ID (must be number)", http.StatusBadRequest)
			return
		}

		var teacherFromDB models.Teacher
		// Note: Using 'tx' for QueryRow ensures we are inside the transaction
		err := tx.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM teachers WHERE id = ?", id).
			Scan(&teacherFromDB.ID, &teacherFromDB.FirstName, &teacherFromDB.LastName, &teacherFromDB.Email, &teacherFromDB.Class, &teacherFromDB.Subject)

		if err == sql.ErrNoRows {
			tx.Rollback()
			utils.WriteJSONError(w, nil, fmt.Sprintf("Teacher with ID %d not found", id), http.StatusNotFound)
			return
		} else if err != nil {
			tx.Rollback()
			utils.WriteJSONError(w, err, "Database lookup error", http.StatusInternalServerError)
			return
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
			utils.WriteJSONError(w, err, fmt.Sprintf("Failed to update teacher %d", id), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.WriteJSONError(w, err, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
