package main

import (
	"flag"
	"fmt"
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

	cqbot.AddGroupMessageHandler(func(m *cqbot.GroupMessage) *cqbot.GroupReplyMessage {
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
		return nil
	})

	cqbot.AddGroupMessageHandler(func(m *cqbot.GroupMessage) *cqbot.GroupReplyMessage {
		if m.Message == nil || !strings.Contains(*m.Message, "炮粉") {
			return nil
		}

		s := `select 
                distinct a.user_id as user_id,
                a.nickname as nickname,
                b.num as count
			  from cqbot_group_message_sender a
			  join (
                select 
                a.user_id,
                count(*) as num
                from cqbot_group_message_sender a
                join cqbot_group_message b on a.pk_id = b.sender_id
                where b.message like '%迅速%'
                and b.create_time >= date_sub(curdate(),interval 7 day)
                group by a.user_id
			  ) b on a.user_id = b.user_id
			  order by b.num desc
              limit 3`
		rows, err := db.Queryx(s)
		if err != nil {
			log.Error(err)
			return nil
		}

		var xunsus []Xunsu
		for rows.Next() {
			x := Xunsu{}
			err := rows.StructScan(&x)
			if err != nil {
				log.Error(err)
				return nil
			}
			xunsus = append(xunsus, x)
		}

		if len(xunsus) <= 2 {
			return nil
		}
		log.Println(xunsus)
		return &cqbot.GroupReplyMessage{
			Reply: fmt.Sprintf("七日迅速榜！No.1(%s[%s])%d次, No.2(%s[%s])%d次, No.3(%s[%s])%d次",
				xunsus[0].Nickname, xunsus[0].UserId, xunsus[0].Count,
				xunsus[1].Nickname, xunsus[1].UserId, xunsus[1].Count,
				xunsus[2].Nickname, xunsus[2].UserId, xunsus[2].Count),
		}
	})

	cqbot.Run("0.0.0.0:" + *port)
}
