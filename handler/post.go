package handler

import (
	"backyard/domain"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
)

var sanitizerStrict = bluemonday.StrictPolicy()

func (h *Handler) NewPost(c echo.Context) error {
	idRegexp := regexp.MustCompilePOSIX("^[a-zA-Z0-9-]+$?")
	id := idRegexp.FindString((c.FormValue("id")))
	if len(id) < 36 {
		return fmt.Errorf("invalid post ID")
	}

	title := c.FormValue("title")
	content := c.FormValue("content")
	draft := c.FormValue("draft") == "on"

	if id != "" && title != "" && content != "" {
		userID := getUserID(c, h.JWTSecret)
		if userID == "" {
			return fmt.Errorf("couldn't get UserID in JWT token")
		}
		tx, err := h.DB.BeginTx(context.TODO(), nil)
		if err != nil {
			return fmt.Errorf("error in begin transaction: %v", err)
		}
		stmt, err := h.DB.Prepare("INSERT INTO posts (id, title, content, draft, createdAt, updatedAt) VALUES (?,?,?,?,?,?)")
		if err != nil {
			return fmt.Errorf("error preparing statement in table posts: %v", err)
		}
		_, err = stmt.Exec(id, title, content, draft, time.Now().UTC(), time.Now().UTC())
		if err != nil {
			return fmt.Errorf("error executing statement in table posts: %v", err)
		}

		stmt, err = h.DB.Prepare("INSERT INTO users_posts (user_id, post_id, relation_type, createdAt, updatedAt) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			return fmt.Errorf("error preparing statement in table users_posts: %v", err)
		}

		_, err = stmt.Exec(userID, id, "AUTHOR", time.Now().UTC(), time.Now().UTC())
		if err != nil {
			return fmt.Errorf("error executing statement in table users_posts: %v", err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("error in commit transaction: %v", err)
		}
	}

	return c.Redirect(http.StatusFound, "/")
}

type PostDTO struct {
	ID        string
	Title     string
	Content   template.HTML
	Draft     bool
	Author    string
	CreatedAt string
}

// TODO remove in favor of getUserID
func isLoggedIn(c echo.Context, JWTSecret string) bool {
	if JWTSecret == "" {
		return false
	}

	cookie, err := c.Cookie("Authorization")
	if err == nil {
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			// SigningMethodHMAC implements the HMAC-SHA family of signing methods.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(JWTSecret), nil
		})
		if err != nil {
			return false
		}
		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return true
		}
	}
	return false
}

func getUserID(c echo.Context, JWTSecret string) string {
	if JWTSecret == "" {
		return ""
	}

	cookie, err := c.Cookie("Authorization")
	if err == nil {
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			// SigningMethodHMAC implements the HMAC-SHA family of signing methods.
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(JWTSecret), nil
		})
		if err != nil {
			return ""
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			expiration, ok := claims["expiration"].(float64)
			// check if the token has expired
			if !ok || time.Now().Compare(time.Unix((int64(expiration)), 0)) > 0 {
				return ""
			}

			userID, ok := claims["userID"].(string)
			if ok {
				return userID
			}
		}
	}
	return ""
}

func (h *Handler) GetPosts(c echo.Context) error {
	rows, err := h.DB.Query("SELECT id, title, content, draft, createdAt, updatedAt FROM posts ORDER BY updatedAt DESC")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	posts := []PostDTO{}
	for rows.Next() {
		p := domain.Post{}
		rows.Scan(&p.ID, &p.Title, &p.Content, &p.Draft, &p.CreatedAt, &p.UpdatedAt)

		posts = append(posts, PostDTO{
			ID:        p.ID,
			Title:     sanitizerStrict.Sanitize(p.Title),
			Content:   safeMd(p.Content),
			Draft:     p.Draft,
			CreatedAt: p.CreatedAt.Format(time.DateOnly),
		})
	}

	return c.Render(http.StatusOK, "index.html", struct {
		Posts    []PostDTO
		UUID     string
		LoggedIn bool
	}{
		Posts:    posts,
		UUID:     uuid.NewString(),
		LoggedIn: isLoggedIn(c, h.JWTSecret),
	})
}

func (h *Handler) GetByID(c echo.Context) error {
	idRegexp := regexp.MustCompilePOSIX("^[a-zA-Z0-9-]+$?")
	id := idRegexp.FindString((c.Param("id")))
	if len(id) < 36 {
		return fmt.Errorf("invalid id")
	}

	row := h.DB.QueryRow("SELECT id, title, content, draft, createdAt, updatedAt FROM posts WHERE id = $1", id)

	if row.Err() != nil {
		panic(row.Err().Error())
	}

	p := domain.Post{}
	row.Scan(&p.ID, &p.Title, &p.Content, &p.Draft, &p.CreatedAt, &p.UpdatedAt)

	return c.Render(http.StatusOK, "post-view.html", struct {
		PostDTO
		LoggedIn bool
	}{
		PostDTO{
			ID:        p.ID,
			Title:     sanitizerStrict.Sanitize(p.Title),
			Content:   safeMd(p.Content),
			Draft:     p.Draft,
			Author:    p.Author,
			CreatedAt: p.CreatedAt.Format(time.DateOnly),
		},
		isLoggedIn(c, h.JWTSecret),
	})
}

func (h *Handler) GetEditPostForm(c echo.Context) error {
	idRegexp := regexp.MustCompilePOSIX("^[a-zA-Z0-9-]+$?")
	id := idRegexp.FindString((c.Param("id")))
	if len(id) < 36 {
		return fmt.Errorf("invalid id")
	}

	row := h.DB.QueryRow("SELECT id, title, content, draft, createdAt, updatedAt from posts WHERE id = $1", id)

	if row.Err() != nil {
		panic(row.Err)
	}

	p := domain.Post{}
	row.Scan(&p.ID, &p.Title, &p.Content, &p.Draft, &p.CreatedAt, &p.UpdatedAt)
	return c.Render(http.StatusOK, "post-edit.html", PostDTO{
		ID:      p.ID,
		Title:   p.Title,
		Content: template.HTML(p.Content),
		Draft:   p.Draft,
	})
}

func (h *Handler) EditPost(c echo.Context) error {
	idRegexp := regexp.MustCompilePOSIX("^[a-zA-Z0-9-]+$?")
	id := idRegexp.FindString((c.FormValue("id")))
	if len(id) < 36 {
		return fmt.Errorf("invalid post ID")
	}
	title := c.FormValue("title")
	content := c.FormValue("content")
	draft := c.FormValue("draft") == "on"

	if id != "" && title != "" && content != "" {
		stmt, err := h.DB.Prepare("UPDATE posts SET title = ?, content = ?, draft = ?,updatedAt = ? WHERE id = ?")
		if err != nil {
			panic(err)
		}
		_, err = stmt.Exec(title, content, draft, time.Now().UTC(), id)
		if err != nil {
			panic(err)
		}
	}

	return c.Redirect(http.StatusFound, "/posts/"+id)
}

func mdToHTML(md string) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func safeMd(content string) template.HTML {
	maybeUnsafeHTML := markdown.ToHTML(mdToHTML(content), nil, nil)
	return template.HTML(bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML))
}
