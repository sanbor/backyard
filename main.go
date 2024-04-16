package main

import (
	"backyard/handler"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
	_ "modernc.org/sqlite"
)

type TemplateRegistry struct {
	templates map[string]*template.Template
}

func hasField(v interface{}, name string) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	return rv.FieldByName(name).IsValid()
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("template not found: " + name)
		return err
	}

	return tmpl.ExecuteTemplate(w, "base.html", data)
}

const DEV_ENV = "dev"
const PRO_ENV = "pro"

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = PRO_ENV
	}

	fmt.Println("Running database schema migrations...")
	db, err := setupDB()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No database schema migration ran. Database schema already in latest version")
		} else {
			fmt.Printf("Error during database schema migration: %v", err)
		}
	}
	JWTSecret, err := fetchSecret(env)
	if err != nil {
		panic(err)
	}
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(JWTSecret),
		TokenLookup: "cookie:Authorization",
		Skipper: func(c echo.Context) bool {
			if c.Request().Method == http.MethodGet || c.Request().Method == http.MethodOptions || c.Path() == "/login" || c.Path() == "/signup" {
				return true
			}

			return false
		},
	}))

	h := handler.Handler{
		DB:           db,
		JWTSecret:    JWTSecret,
		EnableSignup: os.Getenv("ENABLE_SIGNUP") == "true",
		Environment:  env,
	}

	// Frontend
	e.GET("/", h.GetPosts)
	e.GET("/posts/:id", h.GetByID)
	e.GET("/posts/:id/edit", h.GetEditPostForm)
	e.GET("/signup", h.GetNewUserForm)
	e.GET("/login", h.GetLoginForm)
	e.GET("/config", h.GetConfigForm)
	e.Static("/static", "assets")
	e.File("/favicon.ico", "assets/favicon.ico")

	t := map[string]*template.Template{
		"index.html":       template.Must(template.New("").Funcs(template.FuncMap{"hasField": hasField}).ParseFiles("templates/index.html", "templates/base.html")),
		"post-view.html":   template.Must(template.New("").Funcs(template.FuncMap{"hasField": hasField}).ParseFiles("templates/post-view.html", "templates/base.html")),
		"post-edit.html":   template.Must(template.New("").Funcs(template.FuncMap{"hasField": hasField}).ParseFiles("templates/post-edit.html", "templates/base.html")),
		"user-login.html":  template.Must(template.New("").Funcs(template.FuncMap{"hasField": hasField}).ParseFiles("templates/user-login.html", "templates/base.html")),
		"user-signup.html": template.Must(template.New("").Funcs(template.FuncMap{"hasField": hasField}).ParseFiles("templates/user-signup.html", "templates/base.html")),
		"config.html":      template.Must(template.New("").Funcs(template.FuncMap{"hasField": hasField}).ParseFiles("templates/config.html", "templates/base.html")),
	}

	e.Renderer = &TemplateRegistry{
		templates: t,
	}

	// Backend
	e.POST("/posts/:id", h.EditPost)
	e.POST("/post", h.NewPost)
	e.POST("/signup", h.NewUser)
	e.POST("/login", h.Login)
	e.POST("/config", h.Config)
	e.GET("/logout", h.Logout)

	// Fancy error pages
	e.HTTPErrorHandler = customHTTPErrorHandler
	addr := os.Getenv("ADDRESS_LISTEN")
	if env == DEV_ENV && addr == "" {
		addr = ":8080"
	}

	tls := false
	if os.Getenv("TLS") == "true" {
		tls = true
	}

	if addr != "" && !tls {
		e.Logger.Fatal(e.Start(addr))
	} else {
		fmt.Println("Listening with TLS enabled")
		// Cache certificates to avoid issues with rate limits (https://letsencrypt.org/docs/rate-limits)
		e.AutoTLSManager.Cache = autocert.DirCache("./.cache")
		if onlyHost := os.Getenv("WHITELIST_HOST"); onlyHost != "" {
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(onlyHost)
		}
		e.Pre(middleware.HTTPSRedirect())
		if addr == "" {
			addr = ":443"
		}
		e.Logger.Fatal(e.StartAutoTLS(addr))
	}
}
func fetchSecret(env string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" && env == DEV_ENV {
		secret = "unsecure"
	}
	if secret == "" {
		return "", errors.New("no secret defined")
	}
	return secret, nil
}
func setupDB() (*sql.DB, error) {
	// PostgresSQL support will come in the future
	dbDriver := os.Getenv("DB_DRIVER")
	dataSourceName := os.Getenv("DB_URL")

	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	var db *sql.DB
	var err error
	var driver database.Driver
	if dbDriver == "sqlite" {
		if dataSourceName == "" {
			dataSourceName = "./backyard.db"
		}
		db, err = sql.Open(dbDriver, dataSourceName+"?_pragma=foreign_keys(1)")
		if err != nil {
			return nil, err
		}
		driver, err = sqlite.WithInstance(db, &sqlite.Config{})
		if err != nil {
			return nil, err
		}
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		dbDriver, driver)
	if err != nil {
		return nil, err
	}

	err = m.Up()

	return db, err
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	if code != http.StatusNotFound {
		c.Logger().Error(err)
	}
	errorPage := fmt.Sprintf("assets/%d.html", code)
	if err := c.File(errorPage); err != nil {
		c.Logger().Error(err)
	}
}
