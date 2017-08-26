package main

import (
	"wxapp-blog/handler"
	. "wxapp-blog/model"
	_ "wxapp-blog/model"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"time"
	"html/template"
)

func main() {
	e := echo.New()
	e.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper: middleware.Skipper(func(c echo.Context) bool {
			if c.Request().Method == "GET" {
				return true
			}
			if c.Path() == "/"+ApiVersion+"/comments" && c.Request().Method == "POST" {
				return true
			}
			return false
		}),
		Validator: func(username, password string, c echo.Context) (bool, error) {
			if username == AppUser && password == AppPassword {
				return true, nil
			}
			return false, nil
		},
	}))
	db := pg.Connect(&pg.Options{
		User:     PgUser,
		Password: PgPassword,
		Addr:     PgAddr,
	})
	defer db.Close()

	t := &Template{
		Templates: template.Must(template.ParseGlob("view/*.html")),
	}
	e.Renderer = t

	h := handler.NewHandler(db)
	h.Tokens = make(map[string]bool)
	go func() {
		for {
			time.Sleep(24 * time.Hour)
			for k := range h.Tokens{
				delete(h.Tokens, k)
			}
		}
	}()
	//Render
	e.GET("/", h.Home)
	e.GET("/articles/:article_id", h.Article)
	e.GET("/contact", h.Contact)
	e.GET("/about", h.About)
	e.GET("/donate", h.Donate)
	//article api
	e.GET("/"+ApiVersion+"/articles", h.ListArticle)
	e.GET("/"+ApiVersion+"/articles/:article_id", h.GetArticle)
	e.POST("/"+ApiVersion+"/articles", h.CreateArticle)
	e.PUT("/"+ApiVersion+"/articles/:article_id", h.UpdateArticle)
	e.DELETE("/"+ApiVersion+"/articles/:article_id", h.DeleteArticle)
	e.PATCH("/"+ApiVersion+"/articles/:article_id", h.LikeArticle)
	//comment api
	//e.GET("/"+ApiVersion+"/articles/:article_id/comments", h.ListComment)
	e.POST("/"+ApiVersion+"/comments", h.CreateComment)
	e.DELETE("/"+ApiVersion+"/comments/:comment_id", h.DeleteComment)
	//topic api
	e.GET("/"+ApiVersion+"/topics", h.ListTopic)
	//e.GET("/"+ApiVersion+"/topics/:topic_id", h.GetTopic)
	e.POST("/"+ApiVersion+"/topics", h.CreateTopic)
	e.PUT("/"+ApiVersion+"/topics/:topic_id", h.UpdateTopic)
	e.DELETE("/"+ApiVersion+"/topics/:topic_id", h.DeleteTopic)
	//admin
	//e.GET("/dashboard", h.AdminInterface)
	logrus.Error(e.Start(":1323"))
}
