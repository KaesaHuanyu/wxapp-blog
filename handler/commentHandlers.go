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
	token := c.FormValue("token")
	if _, ok := h.Tokens[token]; ok {
		delete(h.Tokens, token)
	} else {
		result["status"] = "ERROR"
		result["error"] = fmt.Sprint("Token Failed.")
		result["token"] = token
		result["status_code"] = http.StatusBadRequest
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}

	comment := model.NewComment(comment_id)
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		if err := c.Bind(comment); err != nil {
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Create Comment Bind ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		if comment.AvatarUrl == "" {
			comment.AvatarUrl = Default_Avatar
		}

		if comment.NickName == "" {
			comment.NickName = c.Request().Host
		}

		if comment.Content == "" || comment.ArticleId == 0 {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("Request Body need content and article_id.")
			result["status_code"] = http.StatusBadRequest
			logrus.Errorf("[%s] Request Body need content and article_id", time.Now().String())
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
		article := model.NewArticle(comment.ArticleId)
		err = tx.Model(article).Where("id = ?", article.Id).Select()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Select Article from Postgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		article.CommentCount++
		_, err = tx.Model(article).Set("comment_count = ?", article.CommentCount).
			Where("id = ?", article.Id).
			Update()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Update Article-CommentCount from Postgres ERROR: [%s]",
				time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		tx.Commit()

		result["status"] = "CREATED"
		result["comment"] = comment
		result["status_code"] = http.StatusCreated
		logrus.Infof("[%s] Create Comment id: [ %d ] in Article( %d ) Success",
			time.Now().String(), comment.Id, comment.ArticleId)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case <-messageChan:
		//return c.JSONPretty(http.StatusCreated, message, "    ")
		return c.Redirect(http.StatusFound, fmt.Sprintf("/articles/%d#%d", comment.ArticleId, comment.Id))
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
		err = tx.Model(comment).Where("id = ?", comment.Id).Select()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Select Comment from Postgres ERROR: [%s]", time.Now().String(), err.Error())
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

		article := model.NewArticle(comment.ArticleId)
		err = tx.Model(article).Where("id = ?", article.Id).Select()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Select Article from Postgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		article.CommentCount--
		_, err = tx.Model(article).Set("comment_count = ?", article.CommentCount).
			Where("id = ?", article.Id).
			Update()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["comment"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Update Article-CommentCount from Postgres ERROR: [%s]",
				time.Now().String(), err.Error())
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
