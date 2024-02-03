package handler

import (
	"backyard/domain"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) Config(ctx echo.Context) error {
	userID := getUserID(ctx, h.JWTSecret)
	if userID == "" {
		return ctx.Redirect(http.StatusFound, "/")
	}
	row := h.DB.QueryRow("select config_id from config where active = true and admin_user_id = $1 order by updatedAt desc", userID)
	oldConfigID := "old-id"
	err := row.Scan(&oldConfigID)
	if err != nil {
		return ctx.Redirect(http.StatusFound, "/")
	}
	idRegexp := regexp.MustCompilePOSIX("^[a-zA-Z0-9-]+$?")
	ID := idRegexp.FindString((ctx.Param("id")))
	if len(ID) < 36 {
		return fmt.Errorf("invalid id")
	}
	formTitle := ctx.FormValue("title")
	formDescription := ctx.FormValue("description")
	c := domain.Config{
		ID:          ID,
		Title:       formTitle,
		Description: formDescription,
	}
	tx, err := h.DB.BeginTx(context.TODO(), nil)
	if err != nil {
		return fmt.Errorf("error in begin transaction: %v", err)
	}
	stmt, err := h.DB.Prepare("update config set active = false where config_id = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(oldConfigID)
	if err != nil {
		return err
	}
	stmt, err = h.DB.Prepare(`insert into config
        values (?,?,?,?,?,?,?,?,?,current_timestamp,current_timestamp`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(c.ID, c.Active, c.BackyardVersion, c.Title, c.Description, c.ImageHome, c.Favicon, c.Footer, userID)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error in commit transaction: %v", err)
	}
	return ctx.Redirect(http.StatusFound, "/config")

}

type ConfigDTO struct {
	ID              string
	Title           string
	Description     string
	ImageHome       string
	Favicon         string
	Footer          string
	BackyardVersion string
	Active          bool
	AdminUserID     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (h *Handler) GetConfigForm(ctx echo.Context) error {
	userID := getUserID(ctx, h.JWTSecret)
	l := ctx.Logger()
	l.Info("hello", "userId", userID)
	if userID == "" {
		// return ctx.Redirect(http.StatusFound, "/")
		return fmt.Errorf("user id empty")
	}

	row := h.DB.QueryRow("select config_id, active, backyard_version, title_home, desc_home, image_home, favicon_home, footer_html, admin_user_id, created_at, updated_at from config where active = true and admin_user_id = $1 order by updated_at desc", userID)
	fmt.Println(userID)
	c := domain.Config{}
	err := row.Scan(&c.ID, &c.Active, &c.BackyardVersion, &c.Title, &c.Description, &c.ImageHome, &c.Favicon, &c.Footer, &c.AdminUserID, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		// return ctx.Redirect(http.StatusFound, "/")
		return err
	}
	return ctx.Render(http.StatusOK, "config.html", ConfigDTO{
		ID:              uuid.NewString(),
		Active:          c.Active,
		BackyardVersion: c.BackyardVersion,
		Title:           c.Title,
		Description:     c.Description,
		ImageHome:       c.ImageHome,
		Favicon:         c.Favicon,
		Footer:          c.Footer,
		AdminUserID:     c.AdminUserID,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	})

}
