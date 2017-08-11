package handler

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"wxapp-blog/model"
)

func (h *handler) CreateComment(c echo.Context) error {
	result := make(map[string]interface{})
	t := time.Now()
	result["time"] = t.Format(layout)
	comment_id := int(t.UnixNano())
	result["comment_id"] = comment_id

	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		comment := model.NewComment(comment_id)
		if err := c.Bind(comment); err != nil {
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Create Comment Bind ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		if comment.Content == "" || comment.ArticleId == 0 || comment.NickName == "" || comment.AvatarUrl == "" {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("some message is nil")
			result["status_code"] = http.StatusBadRequest
			logrus.Errorf("[%s] some message is nil", time.Now().String())
			errorChan <- result
			return
		}

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Tx Begin ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		//插入comment
		err = tx.Insert(comment)
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Insert Comment To Postgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "CREATED"
		result["comment"] = comment
		result["status_code"] = http.StatusCreated
		logrus.Infof("[%s] Create Comment id: [ %d ] in Article( id: [ %d ] ) Success", time.Now().String(), comment.Id, comment.ArticleId)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusCreated, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Create Comment timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["comment"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) DeleteComment(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	comment_id, err := strconv.Atoi(c.Param("comment_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["error"] = fmt.Sprint(err)
		result["status_code"] = http.StatusBadRequest
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}
	result["comment_id"] = comment_id

	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		comment := model.NewComment(comment_id)

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Tx Begin ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		//删除comment
		_, err = tx.Model(comment).Where("id = ?", comment.Id).Delete()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Insert Comment To Postgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "OK"
		result["comment"] = comment
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Delete Comment: [ %d ] Success", time.Now().String(), comment.Id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusCreated, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Delete Comment timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}
