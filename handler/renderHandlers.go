package handler

import (
	"github.com/labstack/echo"
	"wxapp-blog/model"
	"strconv"
	"time"
	"net/http"
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/russross/blackfriday"
	"crypto/md5"
	"io"
	"github.com/go-pg/pg/orm"
	"io/ioutil"
	"html/template"
)

func (h *handler) Home(c echo.Context) error {
	result := make(map[string]interface{})
	articles := model.NewArticleSlice()
	topics := model.NewTopicsSlice()
	since := c.QueryParam("since")
	date_limit := getDateLimit(since)
	topic := c.QueryParam("topic")

	var err error
	if topic != "" {
		topic_id, err := strconv.Atoi(topic)
		if err != nil {
			result["articles"] = nil
			result["topics"] = nil
			logrus.Errorf("[%s] strconv.Atoi(topic) ERROR: [%s]", time.Now().String(), err.Error())
			return c.NoContent(http.StatusBadRequest)
		}
		err = h.DB.Model(&articles).Order("id DESC").Where("topic_id = ?", topic_id).
			Where("id > ?", date_limit).Select()
	} else {
		err = h.DB.Model(&articles).Order("id DESC").Where("id > ?", date_limit).
			Select()
	}
	if err != nil {
		result["articles"] = nil
		result["topics"] = nil
		logrus.Errorf("[%s] List Articles ERROR: [%s]", time.Now().String(), err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	result["articles"] = articles
	logrus.Infof("[%s] Get Articles Success", time.Now().String())

	err = h.DB.Model(&topics).Select()
	if err != nil {
		result["articles"] = nil
		result["topics"] = nil
		logrus.Errorf("[%s] List Topics ERROR: [%s]", time.Now().String(), err.Error())
		return err
	}
	result["topics"] = topics
	logrus.Infof("[%s] List Topics Success", time.Now().String())
	return c.Render(http.StatusOK, "home", result)
}

func (h *handler) Article(c echo.Context) error {
	result := make(map[string]interface{})
	article_id, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		result["article"] = nil
		logrus.Errorf("[%s] article_id is nil", time.Now().String())
		return c.NoContent(http.StatusBadRequest)
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		//数据库操作
		article := model.NewArticle(article_id)
		err := h.DB.Model(article).Column("article.*", "Comments").Where("id = ?", article_id).
			Relation("Comments", func(q *orm.Query) (*orm.Query, error) {
			return q, nil
		}).First()
		if err != nil {
			result["article"] = nil
			logrus.Errorf("[%s] Get Article ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		article.ClickCount++
		_, err = h.DB.Model(article).Set("click_count = ?", article.ClickCount).Where("id = ?", article_id).
			Update()
		if err != nil {
			result["article"] = nil
			logrus.Errorf("[%s] Get Article ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		result["content"] = template.HTML(string(blackfriday.MarkdownCommon([]byte(article.Content))))
		result["article"] = article
		logrus.Infof("[%s] Get Article id: [ %d ] Success", time.Now().String(), article.Id)
		messageChan <- result
	}(c, messageChan, errorChan)

	crutime := time.Now().Unix()
	hash := md5.New()
	io.WriteString(hash, strconv.FormatInt(crutime, 10))
	token := fmt.Sprintf("%x", hash.Sum(nil))
	h.Tokens[token] = true
	result["token"] = token
	select {
	case message := <-messageChan:
		return c.Render(http.StatusOK, "article", message)
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Get Article timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["article"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) Contact(c echo.Context) error {
	file, err := ioutil.ReadFile("view/contact.html")
	if err != nil {
		logrus.Errorf("[ %s ] Contact ioutil.ReadFile error: [ %s ]", time.Now().String(), err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.HTML(http.StatusOK, string(file))
}

func (h *handler) About(c echo.Context) error {
	file, err := ioutil.ReadFile("view/about.html")
	if err != nil {
		logrus.Errorf("[ %s ] About ioutil.ReadFile error: [ %s ]", time.Now().String(), err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.HTML(http.StatusOK, string(file))
}

func (h *handler) Donate(c echo.Context) error {
	file, err := ioutil.ReadFile("view/donate.html")
	if err != nil {
		logrus.Errorf("[ %s ] About ioutil.ReadFile error: [ %s ]", time.Now().String(), err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.HTML(http.StatusOK, string(file))
}