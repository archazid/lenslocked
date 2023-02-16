package main

import (
	"fmt"
	"net/http"

	"archazid.io/lenslocked/controllers"
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

	var usersC controllers.Users
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "signup.tmpl"))
	r.Get("/users/new/", usersC.New)
	r.Post("/users/create/", usersC.Create)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
