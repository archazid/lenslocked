package main

import (
	"fmt"
	"net/http"

	"archazid.io/lenslocked/controllers"
	"archazid.io/lenslocked/migrations"
	"archazid.io/lenslocked/models"
	"archazid.io/lenslocked/templates"
	"archazid.io/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func main() {
	// Setup a database connection
	cfg := models.DefaultPostgresConfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// Database migrations
	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Setup our model services
	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}

	// Setup CSRF middleware
	csrfKey := "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	csrfMw := csrf.Protect(
		[]byte(csrfKey),
		// TODO: Fix this before deploying
		csrf.Secure(false),
	)
	// Setup user middleware
	umw := controllers.UserMiddleware{
		SessionService: &sessionService,
	}

	// Setup our controllers
	usersC := controllers.Users{
		UserService:    &userService,
		SessionService: &sessionService,
	}
	usersC.Templates.New = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "signup.tmpl"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "signin.tmpl"))

	// Setup our router
	r := chi.NewRouter()
	// Apply middleware
	r.Use(csrfMw)
	r.Use(umw.SetUser)

	// Setup our routes
	r.Get("/", controllers.StaticHandler(views.Must(
		views.ParseFS(templates.FS, "base.tmpl", "home.tmpl"))))
	r.Get("/contact", controllers.StaticHandler(views.Must(
		views.ParseFS(templates.FS, "base.tmpl", "contact.tmpl"))))
	r.Get("/faq", controllers.FAQ(views.Must(
		views.ParseFS(templates.FS, "base.tmpl", "faq.tmpl"))))

	r.Get("/signup", usersC.New)
	r.Post("/users", usersC.Create)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start our server
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", r)
}
