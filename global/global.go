package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// This is an example of a trivial Web Service running against a PostgreSQL
// database, to illustrate the problem of how to pass sql.DB instances (or
// other pieces of global application state) to your http.HandlerFunc handlers.

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("pgx",
		"postgres://localhost:5432/example?sslmode=disable")
	if err != nil {
		panic(err)
	}

	router := makeRouter()

	err = http.ListenAndServe(":80", router)
	log.Fatal(err)
}

func makeRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/msg/{id}", getMessage)
	r.Put("/msg/{id}", putMessage)
	r.Post("/msg", postMessage)
	return r
}

func getMessage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	row := db.QueryRowContext(r.Context(),
		"select msg from messages where id = $1", id)
	var msg string
	err = row.Scan(&msg)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(msg))
}

func putMessage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	_, err = db.ExecContext(r.Context(),
		"update messages set msg = $2 where id = $1", id, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postMessage(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	row := db.QueryRowContext(r.Context(),
		"insert into messages (msg) values ($1) returning id", string(body))
	var id int64
	err = row.Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(strconv.FormatInt(id, 10)))
}
