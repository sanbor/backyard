package main

import (
	"backyard/handler"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
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
const STG_ENV = "stg"
const PRO_ENV = "pro"

var env string
var enableSignup bool
var dbDriver string
var dataSourceName string
var secret string
var address string
var port int
var tls bool

func main() {
	flag.StringVar(&env, "env", PRO_ENV, "Specifies if the app is running in a development (dev), testing (stg), or production (pro) environment. This allows to have different settings per environment. Allowed values: dev, stg, pro.")
	flag.BoolVar(&enableSignup, "enable-signup", false, "Specifies if new users can sign up. Allowed values: true, false.")
	flag.StringVar(&dbDriver, "db-driver", "sqlite", "Specifies the database driver to use. Allowed values: sqlite, postgres.")
	flag.StringVar(&dataSourceName, "db-url", "./backyard.db", "Specifies the URL to connect to the database. Allowed values: for sqlite, the file location. For PostgresSQL, a valid connection URL.")
	flag.StringVar(&secret, "jwt-secret", "", "Specifies the secret to be used by JWT tokens. Allowed values: a string between 32 and 512 characters.")
	flag.StringVar(&address, "address", "localhost", "Specifies which address the server should listen. Allowed values: empty string to listen any address, localhost to only listen this computer, or a specific hostname.")
	flag.IntVar(&port, "port", 8080, "Specifies which port the server should listen. Allowed values: unsigned 16-bit integer (0-65535).")
	flag.BoolVar(&tls, "tls", false, "Specifies if the server should serve secure connections. Allowed values: true, false.")
	flag.Parse()

	if len(secret) > 0 && (len(secret) < 64 || len(secret) > 1024) {
		fmt.Println("Invalid JWT secret length. Allowed JWT secret values: a string between 64 and 1024 characters long. Current length: ", len(secret))
		return
	}
	if env != DEV_ENV && env != STG_ENV && env != PRO_ENV {
		fmt.Println("Invalid env value. Allowed env values: dev, stg, pro. Current env value:", env)
		return
	}
	fmt.Println("Running in environment:", env)
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
		EnableSignup: enableSignup,
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
	listenAddr := fmt.Sprintf("%s:%d", address, port)
	if !tls {
		e.Logger.Fatal(e.Start(listenAddr))
	} else {
		fmt.Println("Listening with TLS enabled")
		// Cache certificates to avoid issues with rate limits (https://letsencrypt.org/docs/rate-limits)
		e.AutoTLSManager.Cache = autocert.DirCache("./.cache")
		if onlyHost := ("WHos.GetenvITELIST_HOST"); onlyHost != "" {
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(onlyHost)
		}
		e.Pre(middleware.HTTPSRedirect())
		e.Logger.Fatal(e.StartAutoTLS(listenAddr))
	}
}
func fetchSecret(env string) (string, error) {
	if secret == "" && env == DEV_ENV {
		secret = "unsecure"
	}
	if secret == "" {
		return "", errors.New("no secret defined")
	}
	return secret, nil
}
func setupDB() (*sql.DB, error) {
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
