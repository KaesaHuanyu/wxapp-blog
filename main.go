package main

import (
	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"handler"
	_ "model"
)

const (
	apiVersion = "v1"
)

func main() {
	e := echo.New()
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == "huanyu0w0" && password == "3.1415926" {
			return true, nil
		}
		return false, nil
	}))
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "admin",
		Addr:     "119.29.243.98:5432",
	})
	defer db.Close()

	h := handler.NewHandler(db)
	//article api
	e.GET("/" + apiVersion + "/articles", h.ListArticle)
	e.GET("/" + apiVersion + "/articles/:id", h.GetArticle)
	e.POST("/" + apiVersion + "/topics/:topicid/articles", h.CreateArticle)
	e.PUT("/" + apiVersion + "/articles/:id", h.UpdateArticle)
	e.DELETE("/" + apiVersion + "/articles/:id", h.DeleteArticle)
	//e.GET("/v1/articles/search", h.SearchArticle)
	e.GET("/" + apiVersion + "/articles/:id/like", h.LikeArticle)
	//comment api
	e.GET("/" + apiVersion + "/articles/:articleid/comments", h.ListComment)
	e.POST("/" + apiVersion + "/articles/:articleid/comments", h.CreateComment)
	e.DELETE("/" + apiVersion + "/articles/:articleid/comments/:commentid", h.DeleteComment)
	//topic api
	e.GET("/" + apiVersion + "/topics", h.ListTopic)
	e.GET("/" + apiVersion + "/topics/:id", h.GetTopic)
	e.POST("/" + apiVersion + "/topics", h.CreateTopic)
	e.PUT("/" + apiVersion + "/topics/:id", h.UpdateTopic)
	e.DELETE("/" + apiVersion + "/topics/:id", h.DeleteTopic)
	logrus.Error(e.Start(":1323"))
}
