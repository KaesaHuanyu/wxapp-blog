package model

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	layout = "2006年 01月02日 15:04:05"
)

type (
	Article struct {
		Id         int               `json:"id"`
		Title      string            `json:"title" form:"title" sql:",null"`
		Content    string            `json:"content" form:"content" sql:",null"` //html & md
		CreateTime string            `json:"create_time"`
		UpdateTime string            `json:"update_time"`
		ClickCount int               `json:"click_count"`
		LikeReader map[string]string `json:"reader_like"`
		Comments   []*Comment        `json:"comments"`
		TopicId    int               `json:"topic_id"` // 0: 话题不明
		Topic      *Topic
	}
	Comment struct {
		Id         int    `json:"id"`
		Content    string `json:"content" form:"content" sql:",null"`
		CreateTime string `json:"create_time"`
		NickName   string `json:"nick_name" form:"nick_name" sql:",null"`
		AvatarUrl  string `json:"avatar_url" form:"avatar_url" sql:",null"`
		Gender     int    `json:"gender" form:"gender"` //性别 0：未知、1：男、2：女
		ArticleId  int    `json:"article_id"`
		Article    *Article
		//Reader *reader `json:"reader"`
	}
	Topic struct {
		Id         int        `json:"id"`
		TopicName  string     `json:"topic_name" form:"topic_name" sql:",null"`
		LikeCount  int        `json:"like_count"`
		CreateTime string     `json:"create_time"`
		UpdateTime string     `json:"update_time"`
		Articles   []*Article `json:"articles"`
	}
	//reader struct {
	//	Id int `json:"id"`
	//	NickName string `json:"nick_name"`
	//	AvatarUrl string `json:"avatar_url"`
	//	Gender string `json:"gender"` //性别 0：未知、1：男、2：女
	//	Province string `json:"province"`
	//	City string `json:"city"`
	//	Country string `json:"country"`
	//	CommentId int `json:"comment_id"`
	//}
)

func init() {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "admin",
		Addr:     "119.29.243.98:5432",
	})
	defer db.Close()

	article, comment, topic := NewModels()
	err := db.CreateTable(article, &orm.CreateTableOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			panic(err)
		} else {
			logrus.Warn("table articles already exists")
		}
	}
	err = db.CreateTable(comment, &orm.CreateTableOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			panic(err)
		} else {
			logrus.Warn("table comments already exists")
		}
	}
	err = db.CreateTable(topic, &orm.CreateTableOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			panic(err)
		} else {
			logrus.Warn("table topics already exists")
		}
	}
}

func NewModels() (*Article, *Comment, *Topic) {
	return &Article{}, &Comment{}, &Topic{}
}

func NewArticle(id, topicid int, title, content string) *Article {
	return &Article{
		Id:         id,
		Title:      title,
		Content:    content,
		CreateTime: time.Unix(int64(id), 0).Format(layout),
		UpdateTime: time.Unix(int64(id), 0).Format(layout),
		LikeReader: make(map[string]string),
		Comments:   []*Comment{},
		TopicId:    topicid,
	}
}

func NewArticleSlice() []*Article {
	return []*Article{}
}

func NewTopicsSlice() []*Topic {
	return []*Topic{}
}

func NewCommentSlice() []*Comment {
	return []*Comment{}
}

func NewComment(id, articleid, gender int, content, nickname, avatarurl string) *Comment {
	return &Comment{
		Id:         id,
		Content:    content,
		CreateTime: time.Unix(int64(id), 0).Format(layout),
		NickName:   nickname,
		AvatarUrl:  avatarurl,
		Gender:     gender,
		ArticleId:  articleid,
	}
}

func NewTopic(id int, topicname string) *Topic {
	return &Topic{
		Id:         id,
		TopicName:  topicname,
		CreateTime: time.Unix(int64(id), 0).Format(layout),
		UpdateTime: time.Unix(int64(id), 0).Format(layout),
		Articles:   []*Article{},
	}
}

//func NewReader(nickname, avatarurl, province, city, country string, gender int) *reader {
//	g := ""
//	switch gender {
//	case 1:
//		g = "男"
//	case 2:
//		g = "女"
//	default:
//		g = "未知"
//	}
//	return &reader{
//		NickName: nickname,
//		AvatarUrl: avatarurl,
//		Gender: g,
//		Province: province,
//		City: city,
//		Country: country,
//	}
//}
