package handler

import (
	"fmt"
	"github.com/go-pg/pg/orm"
	"github.com/labstack/echo"
	"github.com/russross/blackfriday"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"wxapp-blog/model"
	"crypto/md5"
	"io"
)

func (h *handler) GetArticle(c echo.Context) error {
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
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Get Article ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		article.ClickCount++
		_, err = h.DB.Model(article).Set("click_count = ?", article.ClickCount).Where("id = ?", article_id).
			Update()
		if err != nil {
			result["status"] = "ERROR"
			result["article"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Get Article ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		article.Content = string(blackfriday.MarkdownCommon([]byte(article.Content)))

		result["status"] = "OK"
		result["article"] = article
		result["status_code"] = http.StatusOK
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
	since := c.QueryParam("since")
	date_limit := getDateLimit(since)
	topic := c.QueryParam("topic")

	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})
	go func(messageChan, errorChan chan map[string]interface{}) {
		var count int
		var err error
		if topic != "" {
			topic_id, err := strconv.Atoi(topic)
			if err != nil {
				result["status"] = "ERROR"
				result["count"] = count
				result["articles"] = nil
				result["error"] = fmt.Sprint(err)
				result["status_code"] = http.StatusInternalServerError
				logrus.Errorf("[%s] strconv.Atoi(topic) ERROR: [%s]", time.Now().String(), err.Error())
				errorChan <- result
				return
			}
			count, err = h.DB.Model(&articles).Order("id DESC").Where("topic_id = ?", topic_id).
				Where("id > ?", date_limit).SelectAndCount()
		} else {
			count, err = h.DB.Model(&articles).Order("id DESC").Where("id > ?", date_limit).
				SelectAndCount()
		}
		if err != nil {
			result["status"] = "ERROR"
			result["count"] = count
			result["articles"] = nil
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] List Articles ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		//wg := sync.WaitGroup{}
		//for _, article := range articles {
		//	wg.Add(1)
		//	go func(article *model.Article) {
		//		defer wg.Done()
		//		article.Comments = model.NewCommentSlice()
		//		err := h.DB.Model(article).Where("id = ?", article.Id).Column("article.*", "Comments").
		//			Relation("Comments", func(q *orm.Query) (*orm.Query, error) {
		//			return q, nil
		//		}).First()
		//		if err != nil {
		//			result["status"] = "ERROR"
		//			result["articles"] = nil
		//			result["error"] = fmt.Sprint(err)
		//			result["status_code"] = http.StatusInternalServerError
		//			logrus.Errorf("[%s] Find Comments ERROR: [%s]",time.Now().String(), err.Error())
		//			errorChan <- result
		//			return
		//		}
		//		article.CommentCount = len(article.Comments)
		//	}(article)
		//	article.LikeCount = len(article.LikeReader)
		//}
		//wg.Wait()

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

		if article.TopicId != 0 {
			topic := model.NewTopic(article.TopicId)
			err = tx.Model(topic).Where("id = ?", article.TopicId).Select()
			if err != nil {
				tx.Rollback()
				result["status"] = "ERROR"
				result["article"] = nil
				result["error"] = fmt.Sprint(err)
				result["status_code"] = http.StatusInternalServerError
				logrus.Errorf("[%s] Select Topic-ArticleCount From Postgres ERROR: [%s]",
					time.Now().String(), err.Error())
				errorChan <- result
				return
			}
			topic.ArticleCount++
			_, err = tx.Model(topic).Set("article_count = ?", topic.ArticleCount).Update()
			if err != nil {
				tx.Rollback()
				result["status"] = "ERROR"
				result["article"] = nil
				result["error"] = fmt.Sprint(err)
				result["status_code"] = http.StatusInternalServerError
				logrus.Errorf("[%s] Update Topic-ArticleCount To Postgres ERROR: [%s]",
					time.Now().String(), err.Error())
				errorChan <- result
				return
			}
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
	article_id, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["error"] = fmt.Sprint("Article Id is nil")
		result["status_code"] = http.StatusInternalServerError
		logrus.Errorf("[%s] Article Id is nil", time.Now().String())
		return c.JSONPretty(http.StatusInternalServerError, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		article := model.NewArticle(article_id)
		if err := c.Bind(article); err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Update Article Bind ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		if article.Title == "" && article.Content == "" && article.TopicId == 0 {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("Request Body need title, content and topic_name.")
			result["status_code"] = http.StatusBadRequest
			logrus.Errorf("[%s] Request Body need title, content and topic_name", time.Now().String())
			errorChan <- result
			return
		}
		article.UpdateTime = t.Format(layout)

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] TxBegin ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}

		switch {
		case article.TopicId != 0:
			_, err = tx.Model(article).Set("topic_id = ?", article.TopicId).
				Set("update_time = ?", article.UpdateTime).
				Where("id = ?", article.Id).
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
			fallthrough
		case article.Content != "":
			_, err = tx.Model(article).Set("content = ?", article.Content).
				Set("update_time = ?", article.UpdateTime).
				Where("id = ?", article.Id).
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
			fallthrough
		case article.Title != "":
			_, err = tx.Model(article).Set("title = ?", article.Title).
				Set("update_time = ?", article.UpdateTime).
				Where("id = ?", article.Id).
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
		}
		tx.Commit()

		result["status"] = "OK"
		result["article_new_title"] = article.Title
		result["article_new_content"] = article.Content
		result["article_new_topic"] = article.TopicId
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Update Article Success", time.Now().String())
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
	article_id, err := strconv.Atoi(c.Param("article_id"))
	if err != nil {
		result["status"] = "ERROR"
		result["article"] = nil
		result["error"] = fmt.Sprint("Article Id is nil")
		result["status_code"] = http.StatusBadRequest
		logrus.Errorf("[%s] Article Id is nil", t.String())
		return c.JSONPretty(http.StatusBadRequest, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		article := model.NewArticle(article_id)

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] TxBegin ERROR: [%s]", time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		err = tx.Model(article).Where("id = ?", article.Id).Select()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Select Article from Postgres ERROR: [%s]",
				time.Now().String(), err.Error())
			errorChan <- result
			return
		}
		if article.TopicId != 0 {
			topic := model.NewTopic(article.TopicId)
			err = tx.Model(topic).Where("id = ?", topic.Id).Select()
			if err != nil {
				result["status"] = "ERROR"
				result["error"] = fmt.Sprint(err)
				result["status_code"] = http.StatusInternalServerError
				logrus.Errorf("[%s] Select Topic from Postgres ERROR: [%s]",
					time.Now().String(), err.Error())
				errorChan <- result
				return
			}
			topic.ArticleCount--
			_, err = tx.Model(topic).Set("article_count = ?", topic.ArticleCount).Update()
			if err != nil {
				tx.Rollback()
				result["status"] = "ERROR"
				result["error"] = fmt.Sprint(err)
				result["status_code"] = http.StatusInternalServerError
				logrus.Errorf("[%s] Update Topic-ArticleCount to Postgres ERROR: [%s]",
					time.Now().String(), err.Error())
				errorChan <- result
				return
			}
		}
		_, err = tx.Model(article).Where("id = ?", article.Id).Delete()
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
		_, err = tx.Model(&comments).Where("article_id = ?", article.Id).Delete()
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
		result["article_id"] = article.Id
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Delete Article [%d] Success", time.Now().String(), article.Id)
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
		result["error"] = fmt.Sprint("Article Title is nil")
		result["status_code"] = http.StatusInternalServerError
		logrus.Errorf("[%s] Article Title is nil", t.String())
		return c.JSONPretty(http.StatusInternalServerError, result, "    ")
	}
	messageChan := make(chan map[string]interface{})
	errorChan := make(chan map[string]interface{})

	go func(c echo.Context, messageChan, errorChan chan map[string]interface{}) {
		article := model.NewArticle(article_id)
		message := &struct {
			NickName  string `json:"nick_name" form:"nick_name"`
			AvatarUrl string `json:"avatar_url" form:"avatar_url"`
		}{}
		if err := c.Bind(message); err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Like Article Bind ERROR: [%s]", t.String(), err.Error())
			errorChan <- result
			return
		}

		if message.NickName == "" {
			message.NickName = c.Request().Host
		}

		if message.AvatarUrl == "" {
			message.AvatarUrl = Default_Avatar
		}

		//数据库操作
		//创建事务
		tx, err := h.DB.Begin()
		if err != nil {
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint("TxBegin ERROR: ", err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] TxBegin ERROR: [%s]", t.String(), err.Error())
			errorChan <- result
			return
		}
		//获取article
		err = tx.Model(article).Where("id = ?", article.Id).Select()
		if err != nil {
			tx.Rollback()
			result["status"] = "ERROR"
			result["error"] = fmt.Sprint(err)
			result["status_code"] = http.StatusInternalServerError
			logrus.Errorf("[%s] Select Article From Postgres ERROR: [%s]", t.String(), err.Error())
			errorChan <- result
			return
		}
		if article.LikeReader == nil {
			article.LikeReader = make(map[string]string)
		}
		article.LikeReader[message.NickName] = message.AvatarUrl
		article.LikeCount++
		_, err = tx.Model(article).Set("like_reader = ?", article.LikeReader).
			Set("like_count = ?", article.LikeCount).
			Where("id = ?", article.Id).
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
		result["like_reader"] = article.LikeReader
		result["like_count"] = article.LikeCount
		result["article_id"] = article.Id
		result["status_code"] = http.StatusOK
		logrus.Infof("[%s] Like Article [ %d ] Success", t.String(), article.Id)
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
