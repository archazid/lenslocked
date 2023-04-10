package main

import (
	"fmt"

	"archazid.io/lenslocked/models"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := models.DefaultPostgresConfig()

	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected!")

	us := models.UserService{
		DB: db,
	}
	user, err := us.Create("bob@bob.com", "bob123")
	if err != nil {
		panic(err)
	}
	fmt.Println(user)
}
