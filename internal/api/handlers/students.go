package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"school-api/internal/models"
	"school-api/internal/repositeries/db"
	"school-api/internal/repositeries/repo"
	"school-api/internal/repositeries/sqlconnect"
	"school-api/pkg/utils"
	"strconv"
)

func GetStudentByIdHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {

		utils.Http500(w, err)
		return
	}
	defer db.Close()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {

		utils.Error(w, "Invalid teacher ID", err)
		return
	}

	student, err := repo.FindStudentByID(id, db.DB)

	if student == nil {
		utils.Error(w, "Student not found", nil)
		return
	} else if err != nil {
		utils.Http500(w, err)
		return
	}

	utils.Success(w, "Student fetched successfully", student)
}

func GetStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	filters := map[string]string{
		"class":      r.URL.Query().Get("class"),
		"first_name": r.URL.Query().Get("first_name"),
		"last_name":  r.URL.Query().Get("last_name"),
		"subject":    r.URL.Query().Get("subject"),
		"email":      r.URL.Query().Get("email"),
	}

	search := r.URL.Query().Get("search")
	sort := utils.BuildSort(r, map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	})

	students, err := repo.FindStudent(db.DB, search, filters, sort)

	if err != nil {
		utils.Http500(w, err)
		return
	}

	if students == nil {
		students = []models.Student{}
	}

	utils.SuccessWithCount(w, "Students fetched successfully", len(students), students)

}

func AddStudentHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	var teacher models.Teacher

	if err := json.NewDecoder(r.Body).Decode(&teacher); err != nil {
		utils.Http500(w, err)
		return
	}

	res, err := repo.AddTeacher(db.DB, &teacher)

	if err != nil {
		utils.Http500(w, err)
		return
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		utils.Error(w, "Failed to retrieve ID", err)
		return
	}

	teacher.ID = int(lastId)

	utils.Success(w, "Teacher added successfully", teacher)
}

func UpdateStudentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.Http500(w, err)
		return
	}

	var updateTeacher models.Teacher
	if err := json.NewDecoder(r.Body).Decode(&updateTeacher); err != nil {
		utils.Error(w, "Invalid request body", err)
		return
	}

	db, err := db.New()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	var existingTeacher models.Teacher

	_, err = repo.UpdateTeacher(db.DB, &existingTeacher, &updateTeacher, id)

	if err == sql.ErrNoRows {
		utils.Error(w, "Teacher not found", err)
		return
	} else if err != nil {
		utils.Http500(w, err)
		return
	}

	utils.Success(w, "Teacher updated successfully", updateTeacher)
}

func DeleteStudentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.Http500(w, err)
		return
	}

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM teachers WHERE id = ?", id)
	if err != nil {
		utils.Http500(w, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.Http500(w, err)
		return
	}

	if rowsAffected == 0 {
		utils.Error(w, "Teacher not found", err)
		return
	}

	utils.Success(w, "Teacher deleted successfully", nil)
}

func DeleteMupltipleStudentsHandler(w http.ResponseWriter, r *http.Request) {
	var ids []int
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		utils.Error(w, "Invalid JSON body (expecting array of IDs)", err)
		return
	}

	db, err := db.New()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	err = repo.DeleteMultipleStudents(db.DB, ids)

	if err != nil {
		utils.Error(w, "Something Went wrong", err)
		return
	}

	utils.Success(w, "Students deleted successfully", nil)
}

func PatchMultipleStudentsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	var updates []map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.Error(w, "Invalid JSON format", err)
		return
	}

	err = repo.PatchMultipleStudents(db.DB, nil, updates)
	if err != nil {
		utils.Error(w, "Something went wrong", err)
		return
	}

	utils.Success(w, "Students updated successfully", nil)
}
