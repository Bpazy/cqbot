package main

import (
	"flag"
	"github.com/Bpazy/cqbot/cqbot"
	"github.com/Bpazy/cqbot/id"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

var (
	port *string
	db   *gorm.DB
)

func init() {
	port = flag.String("port", "12345", "port")
	dataSourceName := flag.String("dns", "", "Data source name. [username[:password]@][protocol[(address)]]/dbname")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{})

	db2, err := gorm.Open("mysql", *dataSourceName+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db = db2

	ensureTable()
}

func main() {
	cqbot.AddPrivateMessageHandler(func(message *cqbot.PrivateMessage) {
		message.CID = id.Id()
		if message.Sender != nil {
			senderId := id.Id()
			message.SenderId = senderId
			message.Sender.CID = senderId
		}
		db.Create(message)
	})

	cqbot.AddDiscussMessageHandler(func(message *cqbot.DiscussMessage) {
		message.CID = id.Id()
		if message.Sender != nil {
			senderId := id.Id()
			message.SenderId = senderId
			message.Sender.CID = senderId
		}
		db.Create(message)
	})

	cqbot.AddGroupMessageHandler(func(message *cqbot.GroupMessage) {
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
	})

	cqbot.Run("0.0.0.0:" + *port)
}

func ensureTable() {
	// ensure table prefix
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return "cqbot_" + defaultTableName
	}

	// table name singular
	db.SingularTable(true)

	if !db.HasTable(cqbot.PrivateMessage{}) {
		db.CreateTable(cqbot.PrivateMessage{})
	}
	if !db.HasTable(cqbot.PrivateMessageSender{}) {
		db.CreateTable(cqbot.PrivateMessageSender{})
	}
	if !db.HasTable(cqbot.GroupMessage{}) {
		db.CreateTable(cqbot.GroupMessage{})
	}
	if !db.HasTable(cqbot.GroupMessageSender{}) {
		db.CreateTable(cqbot.GroupMessageSender{})
	}
	if !db.HasTable(cqbot.DiscussMessage{}) {
		db.CreateTable(cqbot.DiscussMessage{})
	}
	if !db.HasTable(cqbot.DiscussMessageSender{}) {
		db.CreateTable(cqbot.DiscussMessageSender{})
	}
	if !db.HasTable(cqbot.GroupMessageAnonymous{}) {
		db.CreateTable(cqbot.GroupMessageAnonymous{})
	}
}
