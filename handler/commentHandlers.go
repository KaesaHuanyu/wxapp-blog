package handler

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"model"
	"net/http"
	"strconv"
	"time"
)

func (h *handler) ListComment(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	return c.JSONPretty(http.StatusOK, result, "    ")
}

func (h *handler) CreateComment(c echo.Context) error {
	result := make(map[string]interface{})
	t := time.Now()
	result["time"] = t.Format(layout)
	comment_id := int(t.Unix())
	article_id, err := strconv.Atoi(c.Param("articleid"))
	if err != nil {
		result["comment_id"] = c.Param("commentid")
		result["article_id"] = c.Param("articleid")
		result["status"] = "ERROR"
		result["error"] = fmt.Sprint(err)
		result["comment"] = nil
		result["status_code"] = http.StatusBadRequest
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}
	result["comment_id"] = comment_id
	result["article_id"] = article_id

	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		comment := model.NewComment(comment_id, article_id, 0, "", "", "")
		if err = c.Bind(comment); err != nil {
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("CreateComment.Bind ERROR: ", err)
			errorChan <- result
			return
		}

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("TxBegin ERROR: ", err)
			errorChan <- result
			return
		}
		////查找article
		//article, _, _ := model.NewModels()
		//_, err = tx.Query(article, `SELECT * FROM articles WHERE id = ?`, article_id)
		//if err != nil {
		//	tx.Rollback()
		//	result["status"] = "ERROR"
		//	result["comment"] = nil
		//	result["error"] = fmt.Sprint(err)
		//	result["status_code"] = http.StatusInternalServerError
		//	logrus.Error("QueryArticleToPostgres ERROR: ", err)
		//	errorChan <- result
		//	return
		//}
		////更新article
		//article.Comments = append(article.Comments, comment)
		//err = tx.Update(article)
		//if err != nil {
		//	tx.Rollback()
		//	result["status"] = "ERROR"
		//	result["comment"] = nil
		//	result["error"] = fmt.Sprint(err)
		//	result["status_code"] = http.StatusInternalServerError
		//	logrus.Error("UpdateArticleToPostgres ERROR: ", err)
		//	errorChan <- result
		//	return
		//}
		//插入comment
		err = tx.Insert(comment)
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Error("InsertCommentToPostgres ERROR: ", err)
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "CREATED"
		result["comment"] = comment
		result["status_code"] = http.StatusCreated
		logrus.Infof("Create Comment id: [ %d ] in Article( id: [ %d ] ) Success", comment_id, article_id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusCreated, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Info("CreateComment timeout")
		result["status"] = "TIMEOUT"
		result["comment"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) DeleteComment(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	return c.JSONPretty(http.StatusNoContent, result, "    ")
}
