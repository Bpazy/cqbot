package main

import (
	"flag"
	"fmt"
	"github.com/Bpazy/cqbot/cqbot"
	"github.com/Bpazy/cqbot/id"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	port        *string
	db          *sqlx.DB
	cqbotClient *cqbot.Client
)

func init() {
	port = flag.String("port", "12345", "port")
	dataSourceName := flag.String("dns", "", "Data source name. [username[:password]@][protocol[(address)]]/dbname")
	httpApiAddr := flag.String("haa", "http://127.0.0.1:5700", "Http API address")
	flag.Parse()

	db2, err := sqlx.Open("mysql", *dataSourceName+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db = db2

	cqbotClient = cqbot.NewClient(*httpApiAddr)
}

func main() {
	cqbotClient.AddPrivateMessageHandler(func(m *cqbot.PrivateMessage) {
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

	cqbotClient.AddGroupMessageHandler(func(m *cqbot.GroupMessage) {
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
		return
	})

	cqbotClient.AddGroupMessageHandler(func(m *cqbot.GroupMessage) {
		if m.Message == nil {
			return
		}
		r := regexp.MustCompile("炮粉通报一下七日内【(.+)】榜")
		keywords := r.FindStringSubmatch(*m.Message)
		if len(keywords) < 2 {
			return
		}
		keyword := keywords[1]

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
                where b.message = '{keyword}'
                and b.create_time >= date_sub(curdate(),interval 7 day)
                group by a.user_id
			  ) b on a.user_id = b.user_id
			  order by b.num desc
              limit 5`
		s = strings.Replace(s, "{keyword}", keyword, -1)
		rows, err := db.Queryx(s)
		if err != nil {
			panic(err)
			return
		}

		var keywordInfos []KeywordInfo
		for rows.Next() {
			x := KeywordInfo{}
			err := rows.StructScan(&x)
			if err != nil {
				panic(err)
				return
			}
			keywordInfos = append(keywordInfos, x)
		}

		if len(keywordInfos) == 0 {
			return
		}
		log.Println(keywordInfos)

		reply := "七日" + keyword + "榜！\r\n"
		template := "No.%d(%s[%s])%d次"
		for index, xunsu := range keywordInfos {
			reply = reply + fmt.Sprintf(template, index+1, xunsu.Nickname, xunsu.UserId, xunsu.Count) + "\r\n"
		}

		cqbotClient.SendMessage(reply, *m.GroupId)
	})

	cqbotClient.AddGroupMessageInterceptor(func(m *cqbot.GroupMessage) bool {
		if !strings.Contains(*m.Message, "炮粉") {
			return false
		}
		requestLimitKey := fmt.Sprintf("cqbot:request:limit:%s:%s", strconv.FormatInt(*m.UserId, 10), *m.Message)
		boolCmd := RedisClient.SetNX(requestLimitKey, 1, 5*time.Second)
		if !boolCmd.Val() {
			cqbotClient.SendMessage("老是喊你爸鸽抹的？泌阳东西子", *m.GroupId)
			return true
		}

		return false
	})

	cqbotClient.Run("0.0.0.0:" + *port)
}
