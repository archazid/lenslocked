package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"archazid.io/lenslocked/controllers"
	"archazid.io/lenslocked/migrations"
	"archazid.io/lenslocked/models"
	"archazid.io/lenslocked/templates"
	"archazid.io/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key    string
		Secure bool
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}

	// TODO: Read the PSQL values from an ENV variable
	cfg.PSQL = models.DefaultPostgresConfig()

	// TODO: SMTP
	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	// TODO: Read the CSRF values from an ENV variable
	cfg.CSRF.Key = "gFvi45R4fy5xNBlnEeZtQbfAVCYEIAUX"
	cfg.CSRF.Secure = false

	// TODO: Read the server values from an ENV variable
	cfg.Server.Address = ":3000"

	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Setup a database connection
	db, err := models.Open(cfg.PSQL)
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
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	emailService := models.NewEmailService(cfg.SMTP)
	pwResetService := &models.PasswordResetService{
		DB: db,
	}

	// Setup CSRF middleware
	csrfMw := csrf.Protect(
		[]byte(cfg.CSRF.Key),
		// TODO: Fix this before deploying
		csrf.Secure(cfg.CSRF.Secure),
	)
	// Setup user middleware
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	// Setup our controllers
	usersC := controllers.Users{
		UserService:          userService,
		SessionService:       sessionService,
		EmailService:         emailService,
		PasswordResetService: pwResetService,
	}
	usersC.Templates.SignUp = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "signup.tmpl"))
	usersC.Templates.SignIn = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "signin.tmpl"))
	usersC.Templates.CheckYourEmail = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "check-your-email.tmpl"))
	usersC.Templates.ForgotPassword = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "forgot-pw.tmpl"))
	usersC.Templates.ResetPassword = views.Must(views.ParseFS(
		templates.FS, "base.tmpl", "reset-pw.tmpl"))

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

	r.Get("/signup", usersC.SignUp)
	r.Post("/signup", usersC.ProcessSignUp)
	r.Get("/signin", usersC.SignIn)
	r.Post("/signin", usersC.ProcessSignIn)
	r.Post("/signout", usersC.ProcessSignOut)
	r.Get("/forgot-pw", usersC.ForgotPassword)
	r.Post("/forgot-pw", usersC.ProcessForgotPassword)
	r.Get("/reset-pw", usersC.ResetPassword)
	r.Post("/reset-pw", usersC.ProcessResetPassword)
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", usersC.CurrentUser)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	// Start the server
	fmt.Printf("Starting the server on %s...\n", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
}
