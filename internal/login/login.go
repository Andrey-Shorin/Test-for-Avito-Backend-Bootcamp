package login

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenStruct struct {
	Token string `json:"token"`
}
type UserCreate struct {
	User_id string `json:"user_id"`
}

func GenerateToken() (string, error) {
	//b := make([]byte, 128)
	b := make([]byte, 5)
	_, err := rand.Read(b)
	if err != nil {

		return "", err
	}
	return hex.EncodeToString(b), nil
}

func DummyLogin(w http.ResponseWriter, r *http.Request) {
	var user_type string = r.FormValue("user_type")
	if !(user_type == "client" || user_type == "moderator") {
		http.Error(w, "Wrong user type", http.StatusBadRequest)
		return
	}
	db, ok := r.Context().Value("db").(*pgxpool.Pool)
	if !ok {
		http.Error(w, "Database connection not found", http.StatusInternalServerError)
		return
	}
	token, err := GenerateToken()
	if err != nil {
		http.Error(w, "Token generation error", http.StatusInternalServerError)
		return
	}

	h := sha1.New()
	h.Write([]byte(token))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	//fmt.Println(sha1_hash)

	_, err = db.Exec(context.Background(), "insert into tokens (token,type,created_at)  values ($1, $2, $3);",
		sha1_hash, user_type, time.Now().UTC())
	if err != nil {
		http.Error(w, "Cant connect to DB : \n"+err.Error(), http.StatusInternalServerError)
		return
	}
	answer := TokenStruct{
		Token: token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}
func Authorization(c context.Context, token string) (string, error) {
	h := sha1.New()
	h.Write([]byte(token))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	db, ok := c.Value("db").(*pgxpool.Pool)
	if !ok {
		return "", errors.New("database connection not found")

	}

	rows, err := db.Query(context.Background(), "select type from  tokens  where token = $1;", sha1_hash)
	if err != nil {
		return "", errors.New("DB Error : \n" + err.Error())

	}
	defer rows.Close()
	var user string
	rows.Next()
	if err := rows.Scan(&user); err != nil {
		return "", errors.New("wrong token")
	}
	return user, nil
}

func Register(w http.ResponseWriter, r *http.Request) {
	var user_type string = r.FormValue("user_type")
	if !(user_type == "client" || user_type == "moderator") {
		http.Error(w, "Wrong user type", http.StatusBadRequest)
		return
	}

	var password = r.FormValue("password")
	if password == "" {
		http.Error(w, "Please provide password", http.StatusBadRequest)
		return
	}
	var email = r.FormValue("email")

	id := uuid.New()
	idString := id.String()
	db, ok := r.Context().Value("db").(*pgxpool.Pool)
	if !ok {
		http.Error(w, "Database connection not found", http.StatusInternalServerError)
		return
	}

	h := sha1.New()
	h.Write([]byte(password))
	sha1_pass := hex.EncodeToString(h.Sum(nil))

	_, err := db.Exec(context.Background(), "insert into users (userID,email,password,type)  values ($1, $2,$3,$4 );",
		idString, email, sha1_pass, user_type)
	if err != nil {
		http.Error(w, "Cant connect to DB : \n"+err.Error(), http.StatusInternalServerError)
		return
	}
	answer := UserCreate{
		User_id: idString,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func Login(w http.ResponseWriter, r *http.Request) {

	var id = r.FormValue("id")
	if id == "" {
		http.Error(w, "Please provide id", http.StatusBadRequest)
		return
	}
	var password = r.FormValue("password")
	if password == "" {
		http.Error(w, "Please provide password", http.StatusBadRequest)
		return
	}

	db, ok := r.Context().Value("db").(*pgxpool.Pool)
	if !ok {
		http.Error(w, "Database connection not found", http.StatusInternalServerError)
		return
	}

	h := sha1.New()
	h.Write([]byte(password))
	sha1_pass := hex.EncodeToString(h.Sum(nil))

	rows, err := db.Query(context.Background(), "SELECT type from users where userID = $1 and password = $2;", id, sha1_pass)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var userType string

	if !rows.Next() {
		http.Error(w, "wrong id or password", http.StatusNotFound)
		return
	}
	if err := rows.Scan(&userType); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := GenerateToken()
	if err != nil {
		http.Error(w, "Token generation error", http.StatusInternalServerError)
		return
	}

	h2 := sha1.New()
	h2.Write([]byte(token))
	sha1_token := hex.EncodeToString(h2.Sum(nil))

	_, err = db.Exec(context.Background(), "delete from  tokens where userID = $1;", id)
	if err != nil {
		http.Error(w, "Cant connect to DB : \n"+err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = db.Exec(context.Background(), "insert into tokens (token,type,created_at,userID)  values ($1, $2, $3,$4);",
		sha1_token, userType, time.Now().UTC(), id)
	if err != nil {
		http.Error(w, "Cant connect to DB : \n"+err.Error(), http.StatusInternalServerError)
		return
	}
	answer := TokenStruct{
		Token: token,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}
