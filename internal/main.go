package main

import (
	"context"
	"fmt"
	"log"
	"main/internal/config"
	"main/internal/flat"
	"main/internal/house"
	"main/internal/login"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("start")
	conf := config.ReadConfig()
	dbpool, err := pgxpool.New(context.Background(), conf.DbURL)
	if err != nil {
		log.Fatal("cant connect to DB")
	}
	defer dbpool.Close()

	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	//r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "db", dbpool)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/dummyLogin", login.DummyLogin)
	r.Post("/house/create", house.HouseCreate)
	r.Post("/flat/create", flat.FlatCreate)
	r.Post("/flat/update", flat.FlatUpdate)
	r.Get("/house/{id:[0-9]+}", house.HouseById)
	r.Post("/register", login.Register)
	r.Post("/login", login.Login)
	r.Post("/house/{id:[0-9]+}/subscribe", house.HouseSubscribe)
	log.Println("Starting server on :8080")

	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		if err := server.Close(); err != nil {
			log.Fatalf("HTTP close error: %v", err)
		}

	}()

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
