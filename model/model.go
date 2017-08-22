package model

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	layout = "2 Jan 2006 15:04:05"
)

var (
	ApiVersion  = "v1"
	PgHost      = "blog.huanyu0w0.cn"
	PgPort      = "5432"
	PgUser      = "postgres"
	PgPassword  = "admin"
	AppUser     = "huanyu0w0"
	AppPassword = "3.1415926"
)

type (
	Article struct {
		Id           int               `json:"id"`
		Title        string            `json:"title" form:"title" sql:",unique,notnull"`
		Content      string            `json:"content" form:"content" sql:",notnull"` //html & md
		CreateTime   string            `json:"create_time" sql:",notnull"`
		UpdateTime   string            `json:"update_time" sql:",notnull"`
		ClickCount   int               `json:"click_count" sql:",notnull"`
		LikeReader   map[string]string `json:"reader_like" sql:",notnull"`
		LikeCount    int               `json:"like_count" sql:",notnull"`
		Comments     []*Comment        `json:"comments"`
		CommentCount int               `json:"comment_count" sql:",notnull"`
		TopicId      int               `json:"topic_id" form:"topic_id" sql:",notnull"` // 0: 话题不明
		Topic        *Topic
	}
	Comment struct {
		Id         int    `json:"id"`
		Content    string `json:"content" form:"content" sql:",notnull"`
		CreateTime string `json:"create_time" sql:",notnull"`
		NickName   string `json:"nick_name" form:"nick_name" sql:",notnull"`
		AvatarUrl  string `json:"avatar_url" form:"avatar_url" sql:",notnull"`
		Gender     int    `json:"gender" form:"gender" sql:",notnull"` //性别 0：未知、1：男、2：女
		ArticleId  int    `json:"article_id" form:"article_id" sql:",notnull"`
		Article    *Article
		//Reader *reader `json:"reader"`
	}
	Topic struct {
		Id        int    `json:"id"`
		TopicName string `json:"topic_name" form:"topic_name" sql:",unique,notnull"`
		//LikeCount  int        `json:"like_count" sql:",notnull"`
		CreateTime   string     `json:"create_time" sql:",notnull"`
		UpdateTime   string     `json:"update_time" sql:",notnull"`
		ArticleCount int        `json:"article_count" sql:",notnull"`
		Articles     []*Article `json:"articles"`
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
		Addr:     "blog.huanyu0w0.cn:5432",
	})
	defer db.Close()

	article, comment, topic := NewArticle(0), NewComment(0), NewTopic(0)
	err := db.CreateTable(article, &orm.CreateTableOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			panic(err)
		} else {
			logrus.Warnf("[%s] table articles already exists", time.Now().String())
		}
	}
	err = db.CreateTable(comment, &orm.CreateTableOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			panic(err)
		} else {
			logrus.Warnf("[%s] table comments already exists", time.Now().String())
		}
	}
	err = db.CreateTable(topic, &orm.CreateTableOptions{})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			panic(err)
		} else {
			logrus.Warnf("[%s] table topics already exists", time.Now().String())
		}
	}
}

func NewArticle(id int) *Article {
	return &Article{
		Id:         id,
		CreateTime: time.Unix(0, int64(id)).Format(layout),
		UpdateTime: time.Unix(0, int64(id)).Format(layout),
		LikeReader: make(map[string]string),
		Comments:   []*Comment{},
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

func NewComment(id int) *Comment {
	return &Comment{
		Id:         id,
		CreateTime: time.Unix(0, int64(id)).Format(layout),
	}
}

func NewTopic(id int) *Topic {
	return &Topic{
		Id:         id,
		CreateTime: time.Unix(0, int64(id)).Format(layout),
		UpdateTime: time.Unix(0, int64(id)).Format(layout),
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
