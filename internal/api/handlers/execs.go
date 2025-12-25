package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"school-api/internal/models"
	"school-api/internal/repositeries/db"
	"school-api/internal/repositeries/repo"
	"school-api/internal/repositeries/sqlconnect"
	"school-api/pkg/utils"
	"strconv"
	"time"
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

	encodedHash, err := repo.EncryptPassword(exec.Password)

	if err != nil {
		utils.Error(w, "error hashing password", err)
		return
	}

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

	isVerified, err := repo.VerifyPassword(req.Password, exec.Password)

	if err != nil {
		utils.Error(w, "Error during pass verify", err)
		return
	}

	if !isVerified {
		utils.Error(w, "username or pass is wrong", nil)
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{

		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Unix(0, 0),
	})

	utils.Success(w, "Logged out successfully", nil)

}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")

	userId, err := strconv.Atoi(idStr)

	if err != nil {
		utils.Error(w, "Error parsing id", err)
		return

	}
	var req models.UpdatePasswordRequest

	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		utils.Error(w, "Error  parsing body", err)
		return
	}
	r.Body.Close()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		utils.Error(w, "Password is required", nil)
		return
	}

	db, err := sqlconnect.ConnectDB()

	if err != nil {
		utils.Error(w, "failed to connect db", err)
		return
	}

	defer db.Close()

	var username string
	var userPassword string

	err = db.QueryRow("SELECT username, password FROM execs WHERE id =?", userId).Scan(&username, &userPassword)

	if err != nil {
		utils.Error(w, "User not found", err)
		return
	}

	ok, err := repo.VerifyPassword(req.CurrentPassword, userPassword)

	if err != nil {
		utils.Error(w, "Error verifying password", err)
		return
	}

	if !ok {
		utils.Error(w, "Password does not match", nil)
		return
	}

	hashedPassword, err := repo.EncryptPassword(req.NewPassword)
	if err != nil {
		utils.Error(w, "Error hashing password", err)
		return
	}

	currentTime := time.Now().Format(time.RFC3339)

	_, err = db.Exec("UPDATE execs SET password =? , password_changed_at =? WHERE id= ?", hashedPassword, currentTime, userId)

	if err != nil {
		utils.Error(w, "Error while updating password", err)
		return
	}

	utils.Success(w, "Password updated successfully", nil)

}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {

	var req struct {
		Email string `json:"email"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		utils.Error(w, "Invalid req body", err)
		return
	}
	r.Body.Close()

	db, err := sqlconnect.ConnectDB()

	if err != nil {
		utils.Error(w, "failed to connect db", err)
		return
	}

	defer db.Close()
	var exec models.Exec
	err = db.QueryRow("SELECT id FROM execs WHERE email=?", req.Email).Scan(&exec.ID)

	if err != nil {
		utils.Error(w, "user not found", err)
		return
	}

	duration, err := strconv.Atoi(os.Getenv("RESET_PASSWORD_EXPIRY"))

	if err != nil {
		utils.Error(w, "Error parsing Expiry Time", err)
		return
	}

	mins := time.Duration(duration)

	expiry := time.Now().Add(mins * time.Minute).Format(time.RFC3339)

	tokenByte := make([]byte, 32)
	_, err = rand.Read(tokenByte)

	if err != nil {
		utils.Error(w, "Error parsing token", err)
		return
	}

	token := hex.EncodeToString(tokenByte)

	hashedToken := sha256.Sum256(tokenByte)

	hashedTokenStr := hex.EncodeToString(hashedToken[:])
	fmt.Println(hashedTokenStr, expiry, exec.ID, "expiry")

	_, err = db.Exec("UPDATE execs SET password_reset_token =?, password_token_expires=? WHERE id=?", hashedTokenStr, expiry, exec.ID)

	if err != nil {
		utils.Error(w, "failed to update db", err)
		return
	}

	resetUrl := fmt.Sprintf("/execs/resetpassword/%s", token)

	message := fmt.Sprintf("Reset pass using the following link: %s , you have %d minutes", resetUrl, int(mins))

	utils.Success(w, message, nil)

}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	token := r.PathValue("resetcode")

	type request struct {
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	var req request

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		utils.Error(w, "invalid request body", err)
		return
	}
	if req.NewPassword == "" || req.ConfirmPassword == "" {
		utils.Error(w, "Password is required", nil)
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		utils.Error(w, "Password does not match", nil)
		return
	}

	db, err := sqlconnect.ConnectDB()

	if err != nil {
		utils.Error(w, "failed to connect db", err)
		return
	}

	defer db.Close()

	var user models.Exec
	bytes, err := hex.DecodeString(token)
	if err != nil {
		utils.Error(w, "failed to read token", err)
		return
	}
	hashedToken := sha256.Sum256(bytes)
	hashedTokenString := hex.EncodeToString(hashedToken[:])

	fmt.Println(token, "<<<>>>")

	query := "SELECT id,email FROM execs WHERE password_reset_token=? AND password_token_expires >?"
	err = db.QueryRow(query, hashedTokenString, time.Now().Format(time.RFC3339)).Scan(&user.ID, &user.Email)

	if err != nil {
		utils.Error(w, "Invalid or expired code", err)
		return
	}

	hashedPassword, err := repo.EncryptPassword(req.NewPassword)

	if err != nil {
		utils.Error(w, "Internal error", err)
		return
	}
	updateQuery := "UPDATE execs SET password=? ,password_reset_token =NULL , password_token_expires=NULL ,password_changed_at=? WHERE id =?"

	_, err = db.Exec(updateQuery, hashedPassword, time.Now().Format(time.RFC3339), user.ID)

	if err != nil {
		utils.Error(w, "Internal db error", err)
		return
	}

	utils.Success(w, "Password updated successfully", nil)

}
