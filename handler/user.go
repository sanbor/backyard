package handler

import (
	"backyard/domain"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) Login(c echo.Context) error {

	formUsername := c.FormValue("username")
	formPassword := c.FormValue("password")

	if len(formUsername) == 0 || len(formPassword) == 0 {
		return c.HTML(http.StatusBadRequest, "Bad request")
	}

	user := new(domain.User)
	user.Username = formUsername

	row := h.DB.QueryRow("SELECT ID, password, email FROM users WHERE username = $1", user.Username)
	if row.Err() != nil {
		fmt.Println(row.Err().Error())
		return c.HTML(http.StatusInternalServerError, "Internal server error")
	}

	var storedPassword string
	err := row.Scan(&user.ID, &storedPassword, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.HTML(http.StatusBadRequest, "Wrong username or password")
		}
		return c.HTML(http.StatusInternalServerError, "Internal server error")
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(formPassword))
	if err != nil {
		return err
	}
	if err != nil {
		return c.HTML(http.StatusBadRequest, "Invalid credentials")
	}
	cookie, err := authorizationCookie(user.ID, h.JWTSecret)
	if err != nil {
		return err
	}

	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/")

}
func (h *Handler) NewUser(c echo.Context) error {
	if h.Environment != "dev" && !h.EnableSignup {
		return c.HTML(http.StatusForbidden, "<h1>Forbidden!</h1><p>Sign up has been disabled.</p>")
	}

	user := domain.User{
		ID:       uuid.NewString(),
		Username: c.FormValue("username"),
	}

	row := h.DB.QueryRow("SELECT COUNT(username) as count FROM users WHERE username = $1", user.Username)
	if row.Err() != nil {
		panic(row.Err().Error())
	}

	var count int
	row.Scan(&count)
	if count != 0 {
		return c.HTML(http.StatusConflict, "Username already taken")
	}

	password := c.FormValue("password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result, err := h.DB.Exec("INSERT INTO users (id, username, password, createdAt, updatedAt) VALUES ($1, $2, $3, $4, $5)", user.ID, user.Username, hashedPassword, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return c.HTML(http.StatusInternalServerError, "User not created")
	}

	cookie, err := authorizationCookie(user.ID, h.JWTSecret)
	if err != nil {
		return err
	}

	c.SetCookie(cookie)

	return c.Redirect(http.StatusFound, "/")
}
func (h *Handler) Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "Authorization"
	cookie.Value = ""
	cookie.Path = "/"

	cookie.Expires = time.Now().Add(-1 * time.Second)
	c.SetCookie(cookie)
	return c.Redirect(http.StatusFound, "/")
}
func (h *Handler) GetNewUserForm(c echo.Context) error {
	return c.Render(http.StatusOK, "user-signup.html", nil)
}
func (h *Handler) GetLoginForm(c echo.Context) error {
	return c.Render(http.StatusOK, "user-login.html", nil)
}

func authorizationCookie(ID string, secret string) (*http.Cookie, error) {
	if secret == "" {
		return nil, errors.New("missing secret")
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["ID"] = ID
	exp := time.Now().Add(time.Hour * 24 * 7)
	claims["expiratoin"] = exp.Unix()
	signedData, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	cookie := new(http.Cookie)
	cookie.Name = "Authorization"
	cookie.Value = signedData
	cookie.Expires = exp
	cookie.Path = "/"

	return cookie, nil
}
