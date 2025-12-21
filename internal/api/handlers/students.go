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

	var student models.Student

	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		utils.Http500(w, err)
		return
	}
	exists, _ := repo.CheckStudentExists(db.DB, student.Email, 0)

	if exists {
		utils.Error(w, "Student with provided email already exists", nil)
		return
	}

	res, err := repo.AddStudent(db.DB, &student)

	if err != nil {
		utils.Http500(w, err)
		return
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		utils.Error(w, "Failed to retrieve ID", err)
		return
	}

	student.ID = int(lastId)

	utils.Success(w, "Student added successfully", student)
}

func UpdateStudentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.Http500(w, err)
		return
	}

	var updateStudent models.Student
	if err := json.NewDecoder(r.Body).Decode(&updateStudent); err != nil {
		utils.Error(w, "Invalid request body", err)
		return
	}

	db, err := db.New()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	var existingStudent models.Student

	_, err = repo.UpdateStudent(db.DB, &existingStudent, &updateStudent, id)

	if err == sql.ErrNoRows {
		utils.Error(w, "Teacher not found", err)
		return
	} else if err != nil {
		utils.Http500(w, err)
		return
	}

	utils.Success(w, "Teacher updated successfully", updateStudent)
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

	_, err = repo.CheckStudentExists(db, "", id)

	if err != nil {
		utils.Error(w, "Student not found2 ", err)
		return
	}

	result, err := db.Exec("DELETE FROM student WHERE id = ?", id)
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
		utils.Error(w, "Student not found", err)
		return
	}

	utils.Success(w, "Teacher deleted successfully", nil)
}

func GetStudentOfTeachers(w http.ResponseWriter, r *http.Request) {

	db, err := sqlconnect.ConnectDB()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	teacherIdStr := r.URL.Query().Get("teacher_id")

	if teacherIdStr == "" {
		utils.Error(w, "teacher_id query param is required", nil)
		return
	}

	teacherId, err := strconv.Atoi(teacherIdStr)

	if err != nil {
		utils.Error(w, "Invalid teacher id string", err)
		return
	}

	var teacherClassId int

	err = db.QueryRow("SELECT class_id FROM teachers WHERE id=? ", teacherId).Scan(&teacherClassId)

	if err == sql.ErrNoRows {
		utils.Error(w, "No teacher found ", err)
		return
	}

	var students []models.Student

	rows, err := db.Query("SELECT id,class_id,first_name,last_name ,email FROM student WHERE class_id=?", teacherClassId)
	if err != nil {

		utils.Error(w, "Student data not found", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var s models.Student

		err = rows.Scan(&s.ID, &s.ClassId, &s.FirstName, &s.LastName, &s.Email)

		students = append(students, s)

	}
	utils.Success(w, "students found successfully", students)

}
