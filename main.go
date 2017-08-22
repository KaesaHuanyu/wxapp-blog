package main

import (
	"fmt"
	"os"
	"wxapp-blog/handler"
	. "wxapp-blog/model"
	_ "wxapp-blog/model"

	"github.com/go-pg/pg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	// err := getEnv()
	// if err != nil {
	// 	logrus.Errorf("[%s] error: [%s]", time.Now().String(), err.Error())
	// 	return
	// }
	e := echo.New()
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == AppUser && password == AppPassword {
			return true, nil
		}
		return false, nil
	}))
	db := pg.Connect(&pg.Options{
		User:     PgUser,
		Password: PgPassword,
		Addr:     PgHost + ":" + PgPort,
	})
	defer db.Close()

	h := handler.NewHandler(db)
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
	e.GET("/dashboard", h.AdminInterface)
	logrus.Error(e.Start(":1323"))
}

func getEnv() error {
	ApiVersion = os.Getenv("apiVersion")
	PgHost = os.Getenv("pg_host")
	PgPort = os.Getenv("pg_port")
	PgUser = os.Getenv("pg_user")
	PgPassword = os.Getenv("pg_password")
	AppUser = os.Getenv("app_user")
	AppPassword = os.Getenv("app_password")
	if ApiVersion == "" || PgHost == "" || PgPort == "" || PgUser == "" ||
		PgPassword == "" || AppUser == "" || AppPassword == "" {
		return fmt.Errorf("SOME ENV is NULL")
	}
	return nil
}
