package handlers

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
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
	"time"

	"golang.org/x/crypto/argon2"
)

func GetExecByIdHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {

		utils.Http500(w, err)
		return
	}
	defer db.Close()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {

		utils.Error(w, "Invalid exec ID", err)
		return
	}

	exec, err := repo.FindExecByID(id, db.DB)

	if exec == nil {
		utils.Error(w, "Exec not found", nil)
		return
	} else if err != nil {
		utils.Http500(w, err)
		return
	}

	utils.Success(w, "Exec fetched successfully", exec)
}

func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	filters := map[string]string{
		"first_name": r.URL.Query().Get("first_name"),
		"last_name":  r.URL.Query().Get("last_name"),
		"email":      r.URL.Query().Get("email"),
		"username":   r.URL.Query().Get("username"),
	}

	search := r.URL.Query().Get("search")
	sort := utils.BuildSort(r, map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"username":   true,
	})

	execs, err := repo.FindExec(db.DB, search, filters, sort)

	if err != nil {
		utils.Http500(w, err)
		return
	}

	if execs == nil {
		execs = []models.Exec{}
	}

	utils.SuccessWithCount(w, "Execs fetched successfully", len(execs), execs)

}

func AddExecHandler(w http.ResponseWriter, r *http.Request) {
	db, err := db.New()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	var exec models.Exec

	if err := json.NewDecoder(r.Body).Decode(&exec); err != nil {
		utils.Http500(w, err)
		return
	}
	exists, _ := repo.CheckExecExists(db.DB, exec.Email, 0)

	if exists {
		utils.Error(w, "Exec with provided email already exists", nil)
		return
	}

	if exec.Password == "" {
		utils.Error(w, "Password is required", nil)
		return
	}

	salt := make([]byte, 16)
	_, err = rand.Read(salt)

	if err != nil {
		utils.Error(w, "Failed to generate salt", err)
		return
	}

	hash := argon2.IDKey([]byte(exec.Password), salt, 1, 64*1024, 4, 32)

	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	exec.Password = encodedHash

	fmt.Println("exec", exec)

	res, err := repo.AddExec(db.DB, &exec)

	if err != nil {
		utils.Http500(w, err)
		return
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		utils.Error(w, "Failed to retrieve ID", err)
		return
	}

	exec.ID = int(lastId)
	exec.Password = ""

	utils.Success(w, "Exec added successfully", exec)
}

func UpdateExecHandler(w http.ResponseWriter, r *http.Request) {
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

func DeleteExecHandler(w http.ResponseWriter, r *http.Request) {
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	var req models.Exec
	var exec = &models.Exec{}

	// data validation

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		utils.Error(w, "Invalid Request", err)
		return
	}

	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		utils.Error(w, "Username and pass is required", nil)
		return
	}

	//search for user
	db, err := sqlconnect.ConnectDB()
	if err != nil {
		utils.Http500(w, err)
		return
	}
	defer db.Close()

	err = db.QueryRow("SELECT id , username, role, password ,inactive FROM execs WHERE username=?", req.Username).Scan(&exec.ID, &exec.Username, &exec.Role, &exec.Password, &exec.Inactive)

	if err == sql.ErrNoRows {
		utils.Error(w, "user not found", err)
		return
	}

	if err != nil {
		utils.Http500(w, err)
		return
	}

	if exec.Inactive {
		utils.Error(w, "user is not active", nil)
		return
	}

	//verify passs

	parts := strings.Split(exec.Password, ".")

	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)

	if err != nil {
		utils.Error(w, "failed to decode pass", err)
		return
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)

	if err != nil {
		utils.Error(w, "failed to decode pass", err)
		return
	}

	hash := argon2.IDKey([]byte(req.Password), salt, 1, 64*1024, 4, 32)

	if len(hash) != len(hashedPassword) {
		utils.Error(w, "wrong username or password", nil)
		return
	}

	if subtle.ConstantTimeCompare(hash, hashedPassword) == 1 {

	} else {
		utils.Error(w, "wrong username or password", nil)
		return
	}

	//genreate token
	token, err := utils.SignToken(exec.ID, req.Username, exec.Role)

	if err != nil {
		utils.Error(w, "JWT error", err)
		return
	}

	// send token as cookie

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	utils.Success(w, "Logged in Successfully", token)

}
