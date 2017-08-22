package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
	"wxapp-blog/model"

	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

func (h *handler) ListTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	topics := model.NewTopicsSlice()

	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(messageChan, errorChan chan map[string]interface{}) {
		count, err := h.DB.Model(&topics).SelectAndCount()
		if err != nil {
			result["status"] = "ERROR"
			result["count"] = count
			result["topics"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] List Topics ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		//wg := sync.WaitGroup{}
		//for _, topic := range topics {
		//	wg.Add(1)
		//	go func(topic *model.Topic) {
		//		defer wg.Done()
		//		topic.Articles = model.NewArticleSlice()
		//		topic.ArticleCount, err = h.DB.Model(&topic.Articles).Where("topic_id = ?", topic.Id).Count()
		//		if err != nil {
		//			result["status"] = "ERROR"
		//			result["count"] = count
		//			result["topics"] = nil
		//			result["error"] = fmt.Sprint(err)
		//			result["status_code"] = http.StatusInternalServerError
		//			logrus.Errorf("[%s] List Topics ERROR: [%s]", time.Now().String(), err.Error())
		//			errorChan <- result
		//			return
		//		}
		//	}(topic)
		//}
		//wg.Wait()

		result["status"] = "OK"
		result["count"] = count
		result["topics"] = topics
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] List Topics Success", time.Now().String())
		messageChan <- result
	}(messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusOK, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] List Topics timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["Topics"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

//暂时不用
func (h *handler) GetTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	id, err := strconv.Atoi(c.Param("topic_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["topic"] = nil
		result["error"] = fmt.Sprint(err)
		result["status_code"] = http.StatusBadRequest
		logrus.Errorf("[%s] Param id ERROR: [%s]", time.Now().String(), err.Error())
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		//数据库操作
		topic := model.NewTopic(id)
		err := h.DB.Model(topic).Column("topic.*", "Articles").Where("id = ?", id).
			Relation("Articles", func(q *orm.Query) (*orm.Query, error) {
				return q, nil
			}).First()
		if err != nil {
			result["status"] = "ERROR"
			result["topic"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Get Topic ERROR: [%s]", time.Now().String(), err.Error())
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
					logrus.Errorf("[%s] Find Comments ERROR: [%s]", time.Now().String(), err.Error())
					errorChan <- result
					return
				}
			}(article)
		}
		wg.Wait()

		result["status"] = "OK"
		result["topic"] = topic
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Get Topic id: [ %d ] Success", time.Now().String(), topic.Id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusOK, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Get Topic timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["article"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) CreateTopic(c echo.Context) error {
	result := make(map[string]interface{})
	t := time.Now()
	result["time"] = t.Format(layout)
	id := int(t.UnixNano())
	result["id"] = id
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		topic := model.NewTopic(id)
		if err := c.Bind(topic); err != nil {
			result["status"] = "ERROR"
			result["topic"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Create Topic Bind ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		if topic.TopicName == "" {
			result["status"] = "ERROR"
			result["topic"] = nil
			result["error"] = fmt.Sprint("topic name is nil")
			result["status_code"] = http.StatusBadRequest
			logrus.Errorf("[%s] topic name is nil", time.Now().String())
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
			logrus.Errorf("[%s] Tx Begin ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		err = tx.Insert(topic)
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["topic"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Insert Topic To Postgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "CREATED"
		result["topic"] = topic
		result["status_code"] = http.StatusCreated
		logrus.Infof("[%s] Create Topic id: [ %d ] Success", time.Now().String(), id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusCreated, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Create Topic timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["topic"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) UpdateTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	topic_id, err := strconv.Atoi(c.Param("topic_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["error"] = fmt.Sprint("Topic Id is nil")
		result["status_code"] = http.StatusBadRequest
		logrus.Errorf("[%s] Topic Id is nil", time.Now().String())
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		topic := model.NewTopic(topic_id)
		if err := c.Bind(topic); err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Update Topic Bind ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		if topic.TopicName == "" {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("topic name is nil")
			result["status_code"] = http.StatusBadRequest
			logrus.Errorf("[%s] topic name is nil", time.Now().String())
			errorChan <- result
			return
		}
		topic.UpdateTime = time.Now().Format(layout)

		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] TxBegin ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		_, err = tx.Model(topic).
			Where("id = ?", topic.Id).
			Set("topic_name = ?", topic.TopicName).
			Set("update_time = ?", topic.UpdateTime).
			Update()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Update Topic ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "OK"
		result["topic_new_name"] = topic.TopicName
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Update Topic Success", time.Now().String())
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusOK, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Update Topic timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["article"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) DeleteTopic(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	topic_id, err := strconv.Atoi(c.Param("topic_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["error"] = fmt.Sprint("Topic Id is nil")
		result["status_code"] = http.StatusBadRequest
		logrus.Errorf("[%s] Topic Id is nil", time.Now().String())
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		topic := model.NewTopic(topic_id)

		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] TxBegin ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		_, err = tx.Model(topic).Where("id = ?", topic.Id).Delete()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Delete Topic ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		articles := model.NewArticleSlice()
		_, err = tx.Model(&articles).
			Set("topic_id = ?", 0).
			Where("topic_id = ?", topic.Id).
			Update()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Delete Topic ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "OK"
		result["status_code"] = http.StatusOK
		logrus.Infof("[&s] Delete Topic id: [ %d ] Success", time.Now().String(), topic.Id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusOK, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Delete Topic timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["article"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}
