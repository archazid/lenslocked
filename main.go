package main

import (
	"fmt"
	"net/http"

	"archazid.io/lenslocked/controllers"
	"archazid.io/lenslocked/models"
	"archazid.io/lenslocked/templates"
	"archazid.io/lenslocked/views"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Get("/", controllers.StaticHandler(views.Must(
		views.ParseFS(templates.FS, "base.tmpl", "home.tmpl"))))
	r.Get("/contact/", controllers.StaticHandler(views.Must(
		views.ParseFS(templates.FS, "base.tmpl", "contact.tmpl"))))
	r.Get("/faq/", controllers.FAQ(views.Must(
		views.ParseFS(templates.FS, "base.tmpl", "faq.tmpl"))))

	// Setup a database connection
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Setup our model services
	userService := models.UserService{
		DB: db,
	}

	// Setup our controllers
	usersC := controllers.Users{
		UserService: &userService,
	}

	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "signup.tmpl"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "signin.tmpl"))
	r.Get("/users/new/", usersC.New)
	r.Post("/users/create/", usersC.Create)
	r.Get("/users/signin/", usersC.SignIn)
	r.Post("/users/auth/", usersC.Authenticate)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
