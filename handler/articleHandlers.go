package handler

import (
	"fmt"
	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"wxapp-blog/model"
	"sync"
)

func (h *handler) GetArticle(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	id, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["article"] = nil
		result["error"] = fmt.Sprint(err)
		result["status_code"] = http.StatusInternalServerError
		logrus.Errorf("[%s] Param id ERROR: [%s]", time.Now().String(), err.Error())
		return c.JSONPretty(http.StatusInternalServerError, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		//数据库操作
		article := model.NewArticle(0)
		err := h.DB.Model(article).Column("article.*", "Comments").Where("id = ?", id).
			Relation("Comments", func(q *orm.Query) (*orm.Query, error) {
				return q, nil
			}).First()
		if err != nil {
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Get Article ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		if mdrender, ok := c.Request().Header["mdrender"]; ok {
			if mdrender[0] == "true" {
				//渲染 article.Content
			}
		}

		result["status"] = "OK"
		result["article"] = article
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Get Article id: [ %d ] Success", time.Now().String(), article.Id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusOK, message, "    ")
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

func (h *handler) ListArticle(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)
	articles := model.NewArticleSlice()
	time_limit := 0
	if time, ok := c.Request().Header["Time_limit"]; ok {
		result["time_limit"] = time[0]
		time_limit = getTimeLimit(time[0])
	}
	if topic_id, ok := c.Request().Header["Topic"]; ok {
		result["topic"] = topic_id[0]
	}

	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(messageChan, errorChan chan map[string]interface{}) {
		var count int
		var err error
		if topic_id, ok := result["topic"]; ok {
			count, err = h.DB.Model(&articles).Where("topic_id = ?", topic_id).
				Where("id > ?", time_limit).SelectAndCount()
		} else {
			count, err = h.DB.Model(&articles).Where("id > ?", time_limit).
				SelectAndCount()
		}
		count, err = h.DB.Model(&articles).SelectAndCount()
		if err != nil {
			result["status"] = "ERROR"
			result["count"] = count
			result["articles"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] List Articles ERROR: [%s]",time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		wg := sync.WaitGroup{}
		for _, article := range articles {
			wg.Add(1)
			go func(article *model.Article) {
				defer wg.Done()
				article.Comments = model.NewCommentSlice()
				err := h.DB.Model(article).Where("id = ?", article.Id).Column("article.*", "Comments").
					Relation("Comments", func(q *orm.Query) (*orm.Query, error) {
					return q, nil
				}).First()
				if err != nil {
					result["status"] = "ERROR"
					result["articles"] = nil
					result["error"] = fmt.Sprint(err)
					result["status_code"] = http.StatusInternalServerError
					logrus.Errorf("[%s] Find Comments ERROR: [%s]",time.Now().String(), err.Error())
					errorChan <- result
					return
				}
				article.CommentCount = len(article.Comments)
			}(article)
		}
		wg.Wait()

		if mdrender, ok := c.Request().Header["Mdrender"];ok {
			if mdrender[0] == "true" {
				//渲染 article.Content
				//wg := sync.WaitGroup{}
				//for _, article := range articles {
				//	wg.Add(1)
				//	go func() {
				//		defer wg.Done()
				//	}()
				//}
				//wg.Wait()
			}
		}

		result["status"] = "OK"
		result["count"] = count
		result["articles"] = articles
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Get Articles Success", time.Now().String())
		messageChan <- result
	}(messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusOK, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Get Article timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["articles"] = nil
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) CreateArticle(c echo.Context) error {
	result := make(map[string]interface{})
	t := time.Now()
	result["time"] = t.Format(layout)
	id := int(t.UnixNano())
	result["id"] = id
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		article := model.NewArticle(id)
		if err := c.Bind(article); err != nil {
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusBadRequest
			logrus.Errorf("[%s] Create Article Bind ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		if article.Title == "" || article.Content == "" {
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint("title or content is nil")
			result["status_code"] = http.StatusBadRequest
			logrus.Errorf("[%s] title or content is nil", time.Now().String())
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
			logrus.Errorf("[%s] Tx Begin ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		err = tx.Insert(article)
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Insert Article To Postgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "CREATED"
		result["article"] = article
		result["status_code"] = http.StatusCreated
		logrus.Infof("[%s] Create Article id: [ %d ] Success", time.Now().String(), id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusCreated, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Create Article timeout", time.Now().String())
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
	id, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["error"] = fmt.Sprint(err)
		result["status_code"] = http.StatusInternalServerError
		logrus.Errorf("[%s] Param id ERROR: [%s]", time.Now().String(), err.Error())
		return c.JSONPretty(http.StatusInternalServerError, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		a := model.NewArticle(id)
		if err := c.Bind(a); err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Update Article Bind ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		if a.Title == "" || a.Content == "" {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("title or content is nil")
			result["status_code"] = http.StatusBadRequest
			logrus.Errorf("[%s] title or content is nil", time.Now().String())
			errorChan <- result
			return
		}
		a.UpdateTime = t.Format(layout)

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] TxBegin ERROR: [%s]", time.Now().String() , err.Error())
			errorChan <- result
			return
		}
		_, err = tx.Model(a).Set("content = ?", a.Content).
			Set("title = ?", a.Title).
			Set("topic_id = ?", a.TopicId).
			Set("update_time = ?", a.UpdateTime).
			Where("id = ?", a.Id).
			Update()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] UpdateArticleToPostgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "OK"
		result["article_new_title"] = a.Title
		result["article_new_content"] = a.Content
		result["article_new_topic"] = a.TopicId
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Update Article id: [ %d ] Success", time.Now().String(), id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusCreated, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] CreateArticle timeout", time.Now().String())
		result["status"] = "TIMEOUT"
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) DeleteArticle(c echo.Context) error {
	result := make(map[string]interface{})
	t := time.Now()
	result["time"] = t.Format(layout)
	id, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["article"] = nil
		result["error"] = fmt.Sprint(err)
		result["status_code"] = http.StatusBadRequest
		logrus.Errorf("[%s] Param article_id ERROR: [%s]", t.String(), err.Error())
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		a := model.NewArticle(id)

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] TxBegin ERROR: [%s]", time.Now().String() , err.Error())
			errorChan <- result
			return
		}
		_, err = tx.Model(a).Where("id = ?", a.Id).Delete()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Delete Article from Postgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		comments := model.NewCommentSlice()
		_, err = tx.Model(&comments).Where("article_id = ?", a.Id).Delete()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Delete Comments from Postgres ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "OK"
		result["article_id"] = a.Id
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Delete Article id: [ %d ] Success", time.Now().String(), id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusOK, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Delete Article timeout", t.String())
		result["status"] = "TIMEOUT"
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}

func (h *handler) SearchArticle(c echo.Context) error {
	result := make(map[string]interface{})
	result["time"] = time.Now().Format(layout)

	return c.JSONPretty(http.StatusOK, result, "    ")
}

func (h *handler) LikeArticle(c echo.Context) error {
	result := make(map[string]interface{})
	t := time.Now()
	result["time"] = t.Format(layout)
	article_id, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["article"] = nil
		result["error"] = fmt.Sprint(err)
		result["status_code"] = http.StatusInternalServerError
		logrus.Errorf("[%s] Param id ERROR: [%s]", t.String(), err.Error())
		return c.JSONPretty(http.StatusInternalServerError, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		a := model.NewArticle(article_id)
		message := &struct {
			NickName string `json:"nick_name" form:"nick_name"`
			AvatarUrl string `json:"avatar_url" form:"avatar_url"`
		}{}
		if err := c.Bind(message); err != nil {
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Like Article Bind ERROR: [%s]", t.String(), err.Error())
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
			logrus.Errorf("[%s] TxBegin ERROR: [%s]", t.String() , err.Error())
			errorChan <- result
			return
		}
		//获取article
		err = tx.Model(a).Column("like_reader").Where("id = ?", a.Id).Select()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Select Article From Postgres ERROR: [%s]", t.String(), err.Error())
			errorChan <- result
			return
		}
		if a.LikeReader == nil {
			a.LikeReader = make(map[string]string)
		}
		a.LikeReader[message.NickName] = message.AvatarUrl
		_, err = tx.Model(a).Set("like_reader = ?", a.LikeReader).
			Where("id = ?", a.Id).
			Update()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Update Article To Postgres ERROR: [%s]", t.String(), err.Error())
			errorChan <- result
			return
		}
		tx.Commit()

		result["status"] = "OK"
		result["like_reader"] = a.LikeReader
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Like Article id: [ %d ] Success", t.String(), article_id)
		messageChan <- result
	}(c, messageChan, errorChan)

	select {
	case message := <-messageChan:
		return c.JSONPretty(http.StatusOK, message, "    ")
	case errMessage := <-errorChan:
		return c.JSONPretty(http.StatusInternalServerError, errMessage, "    ")
	case <-time.After(10 * time.Second):
		logrus.Infof("[%s] Like Article timeout", t.String())
		result["status"] = "TIMEOUT"
		result["status_code"] = http.StatusGatewayTimeout
		return c.JSONPretty(http.StatusGatewayTimeout, result, "    ")
	}
}
