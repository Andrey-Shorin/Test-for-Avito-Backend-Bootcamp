package flat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"main/internal/login"
	"main/internal/subscribe"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Flat struct {
	Id       int64  `json:"id"`
	House_id int    `json:"house_id"`
	Price    int    `json:"price"`
	Rooms    int    `json:"rooms"`
	Status   string `json:"status"`
}

func FlatCreate(w http.ResponseWriter, r *http.Request) {
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
	if !(user != "moderator" || user != "client") {
		http.Error(w, "Authorization failure", http.StatusUnauthorized)
		return
	}

	var id string = r.FormValue("house_id")
	idInt, err := strconv.Atoi(id)
	if err != nil || id == "" {
		http.Error(w, "Incorrect id", http.StatusBadRequest)
		return
	}

	var price = r.FormValue("price")
	if price == "" {
		http.Error(w, "Please provide price", http.StatusBadRequest)
		return
	}
	priceInt, err := strconv.Atoi(price)
	if err != nil {
		http.Error(w, "Incorrect price", http.StatusBadRequest)
		return
	}

	var room = r.FormValue("room")
	if room == "" {
		http.Error(w, "Please provide room", http.StatusBadRequest)
		return
	}
	roomInt, err := strconv.Atoi(room)
	if err != nil {
		http.Error(w, "Incorrect room", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(context.Background(), "Insert into flats (houseId, price, rooms, status) values ($1,$2,$3,$4)"+
		"returning flatId;", idInt, priceInt, roomInt, "created")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var flatId int64
	rows.Next()
	err = rows.Err()
	if err != nil {
		http.Error(w, "Wrong house id : "+err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(rows.Err())
	if err := rows.Scan(&flatId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = db.Exec(context.Background(), "update houses set (update_at = $1)  where id = $2;", time.Now().UTC(), idInt)
	if err != nil {
		log.Printf("Updating house time error on id = %d \n", idInt)
		log.Print(err.Error())
	}

	answer := Flat{
		Id:       flatId,
		House_id: idInt,
		Price:    priceInt,
		Rooms:    roomInt,
		Status:   "created",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)
}

func FlatUpdate(w http.ResponseWriter, r *http.Request) {
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

	var idString = r.FormValue("id")
	if idString == "" {
		http.Error(w, "Please provide id", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, "Incorrect id", http.StatusBadRequest)
		return
	}

	var price = r.FormValue("price")
	if price == "" {
		http.Error(w, "Please provide price", http.StatusBadRequest)
		return
	}
	priceInt, err := strconv.Atoi(price)
	if err != nil {
		http.Error(w, "Incorrect price", http.StatusBadRequest)
		return
	}

	var room = r.FormValue("room")
	if room == "" {
		http.Error(w, "Please provide room", http.StatusBadRequest)
		return
	}
	roomInt, err := strconv.Atoi(room)
	if err != nil {
		http.Error(w, "Incorrect room", http.StatusBadRequest)
		return
	}

	var status = r.FormValue("status")
	if status == "" {
		rows, err := db.Query(context.Background(), `update  flats  set price = $1 , rooms = $2 where flatID =  $3  returning flatId
		, houseId, price ,rooms, status ;`, priceInt, roomInt, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		answer := Flat{}

		if !rows.Next() {
			http.Error(w, "wrong id ", http.StatusBadRequest)
			return
		}
		if err := rows.Scan(&answer.Id, &answer.House_id, &answer.Price, &answer.Rooms, &answer.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(answer)
		return

	}
	if !(status == "created" || status == "approved" || status == "declined" || status == "on moderation") {
		http.Error(w, "Wrong status", http.StatusBadRequest)
		return
	}
	if status == "on moderation" {
		rows, err := db.Query(context.Background(), `update  flats set status = $1 ,price = $2 , rooms = $3 where flatID =  $4 and status != 'on moderation' returning flatId
		, houseId, price ,rooms, status ;`, status, priceInt, roomInt, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		answer := Flat{}

		if !rows.Next() {
			http.Error(w, "wrong id or already on moderation", http.StatusBadRequest)
			return
		}
		if err := rows.Scan(&answer.Id, &answer.House_id, &answer.Price, &answer.Rooms, &answer.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(answer)
		return
	}
	rows, err := db.Query(context.Background(), `update  flats set  status = $1 ,price = $2 , rooms = $3 where flatID =  $4  returning flatId
		, houseId, price ,rooms, status ;`, status, priceInt, roomInt, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	answer := Flat{}

	if !rows.Next() {
		http.Error(w, "wrong id", http.StatusBadRequest)
		return
	}
	if err := rows.Scan(&answer.Id, &answer.House_id, &answer.Price, &answer.Rooms, &answer.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if answer.Status == "approved" {
		go subscribe.Notify(db, answer.House_id, answer.Id)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answer)

}
