package handler

import (
	"fmt"
	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"wxapp-blog/model"
)

func (h *handler) ListTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	return c.JSONPretty(http.StatusOK, result, "    ")
}

func (h *handler) GetTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		//数据库操作
		_, _, topic := model.NewModels()
		err := h.DB.Model(topic).Column("topic.*", "Articles").
			Relation("Articles", func(q *orm.Query) (*orm.Query, error) {
				return q, nil
			}).First()
		if err != nil {
			result["status"] = "ERROR"
			result["topic"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("GetTopic ERROR: ", err)
			errorChan <- result
			return
		}
		wg := sync.WaitGroup{}
		for _, article := range topic.Articles {
			wg.Add(1)
			go func(article *model.Article) {
				defer wg.Done()
				article.Comments = model.NewCommentSlice()
				err := h.DB.Model(article).Column("article.*", "Comments").
					Relation("Comments", func(q *orm.Query) (*orm.Query, error) {
						return q, nil
					}).First()
				if err != nil {
					result["status"] = "ERROR"
					result["topic"] = nil
					result["error"] = fmt.Sprint(err)
					result["status_code"] = http.StatusInternalServerError
					logrus.Error("FindComments ERROR: ", err)
					errorChan <- result
					return
				}
			}(article)
		}
		wg.Wait()

		result["status"] = "SUCCESS"
		result["topic"] = topic
		result["status_code"] = http.StatusOK
		logrus.Infof("Get Article id: [ %d ] Success", topic.Id)
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

func (h *handler) CreateTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	id := int(rand.Int31())
	result["id"] = id
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		topic := model.NewTopic(id, "")
		if err := c.Bind(topic); err != nil {
			result["status"] = "ERROR"
			result["topic"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("CreateTopic.Bind ERROR: ", err)
			errorChan <- result
			return
		}

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["topic"] = nil
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("TxBegin ERROR: ", err)
			errorChan <- result
			return
		}

		//查找topic
		//_, _, articleTopic := model.NewModels()
		//err = tx.Query(articleTopic, `SELECT * FROM topics WHERE id = ?`, topicid)

		//插入article
		err = tx.Insert(topic)
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["topic"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("InsertTopicToPostgres ERROR: ", err)
			errorChan <- result
			return
		}

		tx.Commit()

		result["status"] = "CREATED"
		result["topic"] = topic
		result["status_code"] = http.StatusCreated
		logrus.Infof("Create Topic id: [ %d ] Success", id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusCreated, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Info("CreateTopic timeout")
		result["status"] = "TIMEOUT"
		result["topic"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) UpdateTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	return c.JSONPretty(http.StatusOK, result, "    ")
}

func (h *handler) DeleteTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	return c.JSONPretty(http.StatusNoContent, result, "    ")
}
