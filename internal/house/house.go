package house

import (
	"context"
	"encoding/json"
	"main/internal/flat"
	"main/internal/login"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type House struct {
	Id         int64  `json:"id"`
	Address    string `json:"address"`
	Year       int    `json:"year"`
	Developer  string `json:"developer"`
	Created_at string `json:"created_at"`
	Update_at  string `json:"update_at"`
}
type FlatArray struct {
	Houses []flat.Flat `json:"flats"`
}

func HouseCreate(w http.ResponseWriter, r *http.Request) {
	db, ok := r.Context().Value("db").(*pgxpool.Pool)
	if !ok {
		http.Error(w, "Database connection not found", http.StatusInternalServerError)
		return
	}
	var bearerToken string = r.FormValue("token")
	if bearerToken == "" { //|| len(strings.Split(bearerToken, " ")) != 2 {

		http.Error(w, "Please provide token", http.StatusBadRequest)
		return
	}
	token := bearerToken //strings.Split(bearerToken, " ")[1]
	user, err := login.Authorization(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if user != "moderator" {
		http.Error(w, "Authorization failure", http.StatusUnauthorized)
		return
	}

	var address string = r.FormValue("address")
	if address == "" {
		http.Error(w, "Please provide address", http.StatusBadRequest)
		return
	}
	var yearString = r.FormValue("year")
	if yearString == "" {
		http.Error(w, "Please provide year", http.StatusBadRequest)
		return
	}
	year, err := strconv.Atoi(yearString)
	if err != nil {
		http.Error(w, "Incorrect year", http.StatusBadRequest)
		return
	}

	var developer = r.FormValue("developer")

	created_at := time.Now()
	update_at := time.Now()
	rows, err := db.Query(context.Background(), "Insert into houses (address, year, developer, created_at, update_at) values ($1,$2,$3,$4,$5)"+
		"returning id;", address, year, developer, created_at.UTC(), update_at.UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var id int64
	rows.Next()
	if err := rows.Scan(&id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	answer := House{
		Id:         id,
		Address:    address,
		Year:       year,
		Developer:  developer,
		Created_at: created_at.Format("2006-01-02T15:04:05Z"),
		Update_at:  update_at.Format("2006-01-02T15:04:05Z"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func HouseById(w http.ResponseWriter, r *http.Request) {

	db, ok := r.Context().Value("db").(*pgxpool.Pool)
	if !ok {
		http.Error(w, "Database connection not found", http.StatusInternalServerError)
		return
	}
	var bearerToken string = r.FormValue("token")
	if bearerToken == "" { //|| len(strings.Split(bearerToken, " ")) != 2 {

		http.Error(w, "Please provide token", http.StatusBadRequest)
		return
	}
	idString := chi.URLParam(r, "id")
	if idString == "" {
		http.Error(w, "Please provide id", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Incorrect id", http.StatusBadRequest)
		return
	}
	token := bearerToken //strings.Split(bearerToken, " ")[1]
	user, err := login.Authorization(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if user != "moderator" && user != "client" {
		http.Error(w, "Authorization failure", http.StatusUnauthorized)
		return
	}
	if user == "client" {
		rows, err := db.Query(context.Background(), "Select * from   flats where  houseId = $1 and status = 'approved';",
			id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		answer := []flat.Flat{}
		for rows.Next() {
			var temp flat.Flat
			if err := rows.Scan(&temp.Id, &temp.House_id, &temp.Price, &temp.Rooms, &temp.Status); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			answer = append(answer, temp)
		}
		ret := FlatArray{Houses: answer}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ret)
		return
	}
	if user == "moderator" {
		rows, err := db.Query(context.Background(), "Select * from   flats where  houseId = $1;",
			id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		answer := []flat.Flat{}
		for rows.Next() {
			var temp flat.Flat
			if err := rows.Scan(&temp.Id, &temp.House_id, &temp.Price, &temp.Rooms, &temp.Status); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			answer = append(answer, temp)
		}
		ret := FlatArray{Houses: answer}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ret)
	}

}

func HouseSubscribe(w http.ResponseWriter, r *http.Request) {
	idString := chi.URLParam(r, "id")
	if idString == "" {
		http.Error(w, "Please provide id", http.StatusBadRequest)
		return
	}
	houseId, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Incorrect id", http.StatusBadRequest)
		return
	}

	var email = r.FormValue("email")
	if email == "" {
		http.Error(w, "Please provide email", http.StatusBadRequest)
		return
	}

	db, ok := r.Context().Value("db").(*pgxpool.Pool)
	if !ok {
		http.Error(w, "Database connection not found", http.StatusInternalServerError)
		return
	}
	var bearerToken string = r.FormValue("token")
	if bearerToken == "" { //|| len(strings.Split(bearerToken, " ")) != 2 {
		http.Error(w, "Please provide token", http.StatusBadRequest)
		return
	}
	token := bearerToken //strings.Split(bearerToken, " ")[1]
	_, err = login.Authorization(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_, err = db.Exec(context.Background(), "insert into subscribe (email,houseId)  values ($1, $2);",
		email, houseId)
	if err != nil {
		http.Error(w, "Cant connect to DB : \n"+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

}
