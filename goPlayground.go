package main

import (
	"fmt"
	"github.com/go-pg/pg"
	"model"
)

func main() {
	//t := time.Now().Unix()
	//fmt.Println(t, time.Unix(0, 0))
	//
	//body, _ := ioutil.ReadFile("golang.org/x/net/html")
	//tree, err := html.Parse(strings.NewReader(string(body)))
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("%v", tree)

	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "admin",
		Addr:     "119.29.243.98:5432",
	})
	defer db.Close()

	//topict := int(time.Now().Unix())
	//topic = model.NewTopic(topict, "testTopic")
	//err = db.Insert(topic)
	//if err != nil {
	//	logrus.Error(err)
	//}
	//logrus.Info(topic)
	//
	//article = model.NewArticle(int(time.Now().Unix()), topict, "testTitle", "testContent")
	//err = db.Insert(article)
	//if err != nil {
	//	logrus.Error(err)
	//}
	//logrus.Info(article)
	//
	//comment = model.NewComment(int(time.Now().Unix()), topict, 1, "testComment",
	//	"huanyu0w0", "")
	//err = db.Insert(comment)
	//if err != nil {
	//	logrus.Error(err)
	//}
	//logrus.Info(comment)

	articlee, _, _ := model.NewModels()
	err := db.Model(articlee).
		Column("article.*", "Topic").
		Where("article.id = ?", 1501779162).
		Select()
	if err == nil {
		fmt.Println(articlee.Topic)
	}
}
