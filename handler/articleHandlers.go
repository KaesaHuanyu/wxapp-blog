package handler

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"model"
	"net/http"
	"time"
	"strconv"
	"github.com/go-pg/pg/orm"
)

func (h *handler) GetArticle(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		//数据库操作
		article, _, _ := model.NewModels()
		err := h.DB.Model(article).Column("article.*", "Comments").
			Relation("Comments", func(q *orm.Query) (*orm.Query, error) {
				return q, nil
		}).First()
		if err != nil {
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("GetArticle ERROR: ", err)
			errorChan <- result
			return
		}

		result["status"] = "SUCCESS"
		result["article"] = article
		result["status_code"] = http.StatusOK
		logrus.Infof("Get Article id: [ %d ] Success", article.Id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
	return c.JSONPretty(http.StatusOK, message, "    ")
	case errMessage := <-errorChan:
	return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
	logrus.Info("CreateArticle timeout")
	result["status"] = "TIMEOUT"
	result["article"] = nil
	result["status_code"] = http.StatusGatewayTimeout
	return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) ListArticle(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	result["articles"] = model.NewArticleSlice()

	//since := c.Param("since")
	//timePoint := getTimePoint(t)
	//topic := c.Param("topic")

	return c.JSONPretty(http.StatusOK, result, "    ")
}

func (h *handler) CreateArticle(c echo.Context) error {
	result := make(map[string]interface{})
	t := time.Now()
	result["time"] = t.Format(layout)
	id := int(t.Unix())
	topicid, err := strconv.Atoi(c.Param("topicid"))
	if err != nil {
		result["topic_id"] = c.Param("topicid")
		result["article_id"] = c.Param("id")
		result["status"] = "ERROR"
		result["error"] = fmt.Sprint(err)
		result["comment"] = nil
		result["status_code"] = http.StatusBadRequest
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}
	result["id"] = id
	result["topic_id"] = topicid
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		a := model.NewArticle(id, topicid, "", "")
		if err := c.Bind(a); err != nil {
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("CreateArticle.Bind ERROR: ", err)
			errorChan <- result
			return
		}

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("TxBegin ERROR: ", err)
			errorChan <- result
			return
		}
		////查找topic
		//_, _, articleTopic := model.NewModels()
		//_, err = tx.Query(articleTopic, `SELECT * FROM topics WHERE id = ?`, topicid)
		//if err != nil {
		//	tx.Rollback()
		//	result["status"] = "ERROR"
		//	result["article"] = nil
		//	result["error"] = fmt.Sprint("InsertArticleToPostgres ERROR: ", err)
		//	result["status_code"] = http.StatusInternalServerError
		//	logrus.Error("UpdateTopicToPostgres ERROR: ", err)
		//	errorChan <- result
		//	return
		//}
		////更新topic
		//articleTopic.Articles = append(articleTopic.Articles, a)
		//err = tx.Update(articleTopic)
		//if err != nil {
		//	tx.Rollback()
		//	result["status"] = "ERROR"
		//	result["article"] = nil
		//	result["error"] = fmt.Sprint("InsertArticleToPostgres ERROR: ", err)
		//	result["status_code"] = http.StatusInternalServerError
		//	logrus.Error("UpdateTopicToPostgres ERROR: ", err)
		//	errorChan <- result
		//	return
		//}
		//插入article
		err = tx.Insert(a)
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("InsertArticleToPostgres ERROR: ", err)
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "CREATED"
		result["article"] = a
		result["status_code"] = http.StatusCreated
		logrus.Infof("Create Article id: [ %d ] Success", id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusCreated, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Info("CreateArticle timeout")
		result["status"] = "TIMEOUT"
		result["article"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) UpdateArticle(c echo.Context) error {
	result := make(map[string]interface{})
	t := time.Now()
	result["time"] = t.Format(layout)
	result["method"] = c.Request().Method
	result["path"] = c.Path()
	result["id"] = c.Param("id")

	return c.JSONPretty(http.StatusOK, result, "    ")
}

func (h *handler) DeleteArticle(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	result["method"] = c.Request().Method
	result["path"] = c.Path()
	result["id"] = c.Param("id")

	return c.JSONPretty(http.StatusNoContent, result, "    ")
}

func (h *handler) SearchArticle(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	return c.JSONPretty(http.StatusOK, result, "    ")
}

func (h *handler) LikeArticle(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	return c.JSONPretty(http.StatusOK, result, "    ")
}
