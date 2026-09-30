package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"backend/api"
	"backend/storage"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

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

	jwtsecret := os.Getenv("JWT_SECRET")
	if jwtsecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	app := api.NewAPI(store, []byte(jwtsecret))

	mux := http.NewServeMux()
	app.RegisterRoutes(mux)

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))

}
