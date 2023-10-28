package handler

import (
	"backyard/domain"
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
	if id != "" && title != "" && content != "" {
		stmt, err := h.DB.Prepare("INSERT INTO posts (id, title, content, createdAt, updatedAt) VALUES (?,?,?,?,?)")
		if err != nil {
			panic(err)
		}
		_, err = stmt.Exec(id, title, content, time.Now().UTC(), time.Now().UTC())
		if err != nil {
			panic(err)
		}
	}

	return c.Redirect(http.StatusFound, "/")
}

type PostDTO struct {
	ID        string
	Title     string
	Content   template.HTML
	Author    string
	CreatedAt string
}

func (h *Handler) GetPosts(c echo.Context) error {
	rows, err := h.DB.Query("SELECT id, title, content, createdAt, updatedAt FROM posts ORDER BY updatedAt DESC")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	posts := []PostDTO{}
	for rows.Next() {
		p := domain.Post{}
		rows.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt)

		posts = append(posts, PostDTO{
			ID:        p.ID,
			Title:     sanitizerStrict.Sanitize(p.Title),
			Content:   safeMd(p.Content),
			CreatedAt: p.CreatedAt.Format(time.DateOnly),
		})
	}

	loggedIn := false
	hmacSampleSecret := []byte("secret")
	cookie, err := c.Cookie("Authorization")
	if err == nil {
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return hmacSampleSecret, nil
		})

		if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			loggedIn = true
		} else {
			fmt.Println(err)
		}
	}
	return c.Render(http.StatusOK, "index.html", struct {
		Posts    []PostDTO
		UUID     string
		LoggedIn bool
	}{
		Posts:    posts,
		UUID:     uuid.NewString(),
		LoggedIn: loggedIn,
	})
}

func (h *Handler) GetByID(c echo.Context) error {
	idRegexp := regexp.MustCompilePOSIX("^[a-zA-Z0-9-]+$?")
	id := idRegexp.FindString((c.Param("id")))
	if len(id) < 36 {
		return fmt.Errorf("invalid id")
	}

	row := h.DB.QueryRow("SELECT id, title, content, createdAt, updatedAt FROM posts WHERE id = $1", id)

	if row.Err() != nil {
		panic(row.Err().Error())
	}

	p := domain.Post{}
	row.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt)

	return c.Render(http.StatusOK, "post-view.html", PostDTO{
		ID:        p.ID,
		Title:     sanitizerStrict.Sanitize(p.Title),
		Content:   safeMd(p.Content),
		Author:    p.Author,
		CreatedAt: p.CreatedAt.Format(time.DateOnly),
	})
}

func (h *Handler) GetEditPostForm(c echo.Context) error {
	idRegexp := regexp.MustCompilePOSIX("^[a-zA-Z0-9-]+$?")
	id := idRegexp.FindString((c.Param("id")))
	if len(id) < 36 {
		return fmt.Errorf("invalid id")
	}

	row := h.DB.QueryRow("SELECT id, title, content, createdAt, updatedAt from posts WHERE id = $1", id)

	if row.Err() != nil {
		panic(row.Err)
	}

	p := domain.Post{}
	row.Scan(&p.ID, &p.Title, &p.Content, &p.CreatedAt, &p.UpdatedAt)
	return c.Render(http.StatusOK, "post-edit.html", PostDTO{
		ID:      p.ID,
		Title:   p.Title,
		Content: template.HTML(p.Content),
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
	if id != "" && title != "" && content != "" {
		stmt, err := h.DB.Prepare("UPDATE posts SET title = ?, content = ?, updatedAt = ? WHERE id = ?")
		if err != nil {
			panic(err)
		}
		_, err = stmt.Exec(title, content, time.Now().UTC(), id)
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
