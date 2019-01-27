package main

import (
	"encoding/json"
	"flag"
	"github.com/Bpazy/cqbot/id"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
)

func main() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/", func(c *gin.Context) {
		err := dispatchMsg(c)
		if err != nil {
			panic(err)
		}
		c.JSON(200, nil)
	})
	_ = r.Run("0.0.0.0:12345") // listen and serve on 0.0.0.0:8080
}

type Model struct {
	CID       string `gorm:"primary_key;not null;varchar(20)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PostType struct {
	PostType string `json:"post_type"` // possible value: message,notice,request
}

type MessagePostType struct {
	PostType
	MessageType string `json:"message_type"`
}

type PrivateMessage struct {
	Model
	PostType    string                `json:"post_type"`    // possible value: message
	MessageType string                `json:"message_type"` // possible value: private
	SubType     string                `json:"sub_type"`     // possible value: friend,group,discuss,other
	MessageId   int32                 `json:"message_id"`
	UserId      int64                 `json:"user_id"`
	Message     string                `json:"message" gorm:"type:text"`
	RawMessage  string                `json:"raw_message" gorm:"type:text"`
	Font        int32                 `json:"font"`
	Sender      *PrivateMessageSender `json:"sender" gorm:"foreignkey:SenderId;association_foreignkey:CID"`
	SenderId    string
}

type PrivateMessageSender struct {
	Model
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
}

type GroupMessageAnonymous struct {
	Model
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Flag string `json:"flag"`
}

type GroupMessageSender struct {
	Model
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
	Area     string `json:"area"`
	Level    string `json:"level"`
	Role     string `json:"role"`
	Title    string `json:"title"`
}

type GroupMessage struct {
	Model
	PostType    string                 `json:"post_type"`    // possible value: message
	MessageType string                 `json:"message_type"` // possible value: private
	SubType     string                 `json:"sub_type"`     // possible value: friend,group,discuss,other
	MessageId   int32                  `json:"message_id"`
	GroupId     int64                  `json:"group_id"`
	UserId      int64                  `json:"user_id"`
	Anonymous   *GroupMessageAnonymous `json:"anonymous"  gorm:"foreignkey:AnonymousId;association_foreignkey:CID"`
	AnonymousId string
	Message     string              `json:"message" gorm:"type:text"`
	RawMessage  string              `json:"raw_message" gorm:"type:text"`
	Font        int32               `json:"font"`
	Sender      *GroupMessageSender `json:"sender" gorm:"foreignkey:SenderId;association_foreignkey:CID"`
	SenderId    string
}

type DiscussMessageSender struct {
	Model
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
}

type DiscussMessage struct {
	Model
	PostType    string                `json:"post_type"`
	MessageType string                `json:"message_type"`
	MessageId   int32                 `json:"message_id"`
	DiscussId   int64                 `json:"discuss_id"`
	UserId      int64                 `json:"user_id"`
	Message     string                `json:"message" gorm:"type:text"`
	RawMessage  string                `json:"raw_message" gorm:"type:text"`
	Font        int32                 `json:"font"`
	Sender      *DiscussMessageSender `json:"sender" gorm:"foreignkey:SenderId;association_foreignkey:CID"`
	SenderId    string
}

// TODO feature
type Notice struct {
}

// TODO feature
type Request struct {
}

// post type
const (
	private = "private"
	group   = "group"
	discuss = "discuss"
)

func dispatchMsg(c *gin.Context) error {
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	defer c.Request.Body.Close()

	postType := PostType{}
	err = json.Unmarshal(bodyBytes, &postType)
	if err != nil {
		return err
	}

	if postType.PostType == "message" {
		messagePostType := MessagePostType{}
		err = json.Unmarshal(bodyBytes, &messagePostType)
		if err != nil {
			return err
		}

		if messagePostType.MessageType == private {
			err = handlePrivateMessage(bodyBytes)
		}
		if messagePostType.MessageType == group {
			err = handleGroupMessage(bodyBytes)
		}
		if messagePostType.MessageType == discuss {
			err = handleDiscussMessage(bodyBytes)
		}
	} else {
		log.Printf("%s: %s", postType.PostType, string(bodyBytes))
	}

	if err != nil {
		return err
	}

	return nil
}

func handleDiscussMessage(bytes []byte) error {
	message := DiscussMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", message.MessageType, message.Sender.Nickname, message.UserId, message.Message)
	saveDiscussMessage(&message)
	return nil
}

func saveDiscussMessage(message *DiscussMessage) {
	message.CID = id.Id()
	if message.Sender != nil {
		senderId := id.Id()
		message.SenderId = senderId
		message.Sender.CID = senderId
	}
	db.Create(message)
}

func handleGroupMessage(bytes []byte) error {
	message := GroupMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", message.MessageType, message.Sender.Nickname, message.UserId, message.Message)
	saveGroupMessage(&message)
	return nil
}

func saveGroupMessage(message *GroupMessage) {
	message.CID = id.Id()
	if message.Sender != nil {
		senderId := id.Id()
		message.SenderId = senderId
		message.Sender.CID = senderId
	}
	if message.Anonymous != nil {
		anonymousId := id.Id()
		message.AnonymousId = anonymousId
		message.Anonymous.CID = anonymousId
	}
	db.Create(message)
}

func handlePrivateMessage(bytes []byte) error {
	message := PrivateMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", message.MessageType, message.Sender.Nickname, message.UserId, message.Message)
	savePrivateMessage(&message)
	return nil
}

func savePrivateMessage(message *PrivateMessage) {
	message.CID = id.Id()
	if message.Sender != nil {
		senderId := id.Id()
		message.SenderId = senderId
		message.Sender.CID = senderId
	}
	db.Create(message)
}

var db *gorm.DB

func init() {
	dataSourceName := flag.String("dns", "", "Data source name. [username[:password]@][protocol[(address)]]/dbname")
	flag.Parse()

	ensureLog()

	db2, err := gorm.Open("mysql", *dataSourceName+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db = db2

	ensureTable()
}

func ensureLog() {
	log.SetFormatter(&log.TextFormatter{})
}

func ensureTable() {
	// ensure table prefix
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "cqbot_" + defaultTableName
	}

	// table name singular
	db.SingularTable(true)

	if !db.HasTable(PrivateMessage{}) {
		db.CreateTable(PrivateMessage{})
	}
	if !db.HasTable(PrivateMessageSender{}) {
		db.CreateTable(PrivateMessageSender{})
	}
	if !db.HasTable(GroupMessage{}) {
		db.CreateTable(GroupMessage{})
	}
	if !db.HasTable(GroupMessageSender{}) {
		db.CreateTable(GroupMessageSender{})
	}
	if !db.HasTable(DiscussMessage{}) {
		db.CreateTable(DiscussMessage{})
	}
	if !db.HasTable(DiscussMessageSender{}) {
		db.CreateTable(DiscussMessageSender{})
	}
	if !db.HasTable(GroupMessageAnonymous{}) {
		db.CreateTable(GroupMessageAnonymous{})
	}
}
