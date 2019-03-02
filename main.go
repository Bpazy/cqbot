package main

import (
	"flag"
	"github.com/Bpazy/cqbot/cqbot"
	"github.com/Bpazy/cqbot/id"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
	"strings"
)

var (
	port *string
	db   *sqlx.DB
)

func init() {
	port = flag.String("port", "12345", "port")
	dataSourceName := flag.String("dns", "", "Data source name. [username[:password]@][protocol[(address)]]/dbname")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{})

	db2, err := sqlx.Open("mysql", *dataSourceName+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db = db2
}

func main() {
	cqbot.AddPrivateMessageHandler(func(m *cqbot.PrivateMessage) {
		m.PkId = id.PId()
		if m.Sender != nil {
			senderId := id.PId()
			m.SenderId = senderId
			m.Sender.PkId = senderId
		}

		db.MustExec("insert into cqbot_private_message (pk_id, message_id, sender_id, user_id, post_type, message_type, sub_type, message, raw_message, font) values (?,?,?,?,?,?,?,?,?,?)",
			m.PkId, m.MessageId, m.SenderId, m.UserId, m.PostType, m.MessageType, m.SubType, m.Message, m.RawMessage, m.Font)
		if m.Sender != nil {
			db.MustExec("insert into cqbot_private_message_sender (pk_id, user_id, nickname, sex, age) values (?,?,?,?,?)",
				m.SenderId, m.UserId, m.Sender.Nickname, m.Sender.Sex, m.Sender.Age)
		}
	})

	cqbot.AddGroupMessageHandler(func(m *cqbot.GroupMessage) {
		m.PkId = id.PId()
		if m.Sender != nil {
			senderId := id.PId()
			m.SenderId = senderId
			m.Sender.PkId = senderId
		}
		if m.Anonymous != nil {
			anonymousId := id.PId()
			m.AnonymousId = anonymousId
			m.Anonymous.PkId = anonymousId
		}

		db.MustExec("insert into cqbot_group_message (pk_id, sender_id, message_id, group_id, anonymous_id, user_id, post_type, message_type, sub_type, message, raw_message, font) values (?,?,?,?,?,?,?,?,?,?,?,?)",
			m.PkId, m.SenderId, m.MessageId, m.GroupId, m.AnonymousId, m.UserId, m.PostType, m.MessageType, m.SubType, m.Message, m.RawMessage, m.Font)
		if m.Sender != nil {
			s := m.Sender
			db.MustExec("insert into cqbot_group_message_sender (pk_id, user_id, nickname, card, sex, age, area, level, role, title) values (?,?,?,?,?,?,?,?,?,?)",
				s.PkId, s.UserId, s.Nickname, s.Card, s.Sex, s.Age, s.Area, s.Level, s.Role, s.Title)
		}
		if m.Anonymous != nil {
			a := m.Anonymous
			db.MustExec("insert into cqbot_group_message_anonymous (pk_id, id, name, flag) values (?,?,?,?)",
				a.PkId, a.Id, a.Name, a.Flag)
		}
	})

	cqbot.AddPrivateMessageHandler(func(message *cqbot.PrivateMessage) {
		if strings.Contains(*message.Message, "lalala") {

		}
	})

	cqbot.Run("0.0.0.0:" + *port)
}
