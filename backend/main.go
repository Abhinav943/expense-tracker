package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/api"
	"backend/storage"

	_ "github.com/lib/pq"
)

func main() {
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to the expense database!")

	store := storage.NewStorage(db)
	app := api.NewAPI(store)

	mux := http.NewServeMux()
	app.RegisterRoutes(mux)

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))

}