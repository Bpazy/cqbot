package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/Bpazy/cqbot/cqbot"
	"github.com/Bpazy/cqbot/id"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/nightexcessive/steamid"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	port        *string
	db          *sqlx.DB
	cqbotClient *cqbot.Client
	redisClient *redis.Client
)

func init() {
	port = flag.String("port", "12345", "Port")
	dataSourceName := flag.String("dns", "", "Data source name. [username[:password]@][protocol[(address)]]/dbname")
	httpApiAddr := flag.String("haa", "http://127.0.0.1:5700", "Http API address")
	redisAddr := flag.String("redis", "127.0.0.1:6379", "Redis address")
	flag.Parse()

	db2, err := sqlx.Open("mysql", *dataSourceName+"?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db = db2

	cqbotClient = cqbot.NewClient(*httpApiAddr)
	redisClient = redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: "",
		DB:       0,
	})
	statusCmd := redisClient.Ping()
	if statusCmd.Err() != nil {
		panic(statusCmd.Err())
	}
}

func main() {
	cqbotClient.AddPrivateMessageInterceptor(func(m *cqbot.PrivateMessage) bool {
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
		return false
	})

	cqbotClient.AddGroupMessageHandler(func(c *cqbot.GroupContext) {
		if !strings.HasPrefix(*c.Message.Message, "炮粉通报一下七日内") {
			return
		}
		r := regexp.MustCompile("炮粉通报一下七日内【(.+)】榜")
		keywords := r.FindStringSubmatch(*c.Message.Message)
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
				and b.group_id = {groupId}
                and b.create_time >= date_sub(curdate(),interval 7 day)
                group by a.user_id
			  ) b on a.user_id = b.user_id
			  order by b.num desc
              limit 5`
		s = strings.Replace(s, "{keyword}", keyword, -1)
		s = strings.Replace(s, "{groupId}", strconv.FormatInt(*c.Message.GroupId, 10), -1)
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
			cqbotClient.SendMessage("miu", *c.Message.GroupId)
			return
		}
		log.Println("keywordInfos: ", keywordInfos)

		reply := "七日" + keyword + "榜！\r\n"
		template := "No.%d(%s[%s])%d次"
		for index, xunsu := range keywordInfos {
			reply = reply + fmt.Sprintf(template, index+1, xunsu.Nickname, xunsu.UserId, xunsu.Count) + "\r\n"
		}

		cqbotClient.SendMessage(reply, *c.Message.GroupId)
	})

	cqbotClient.AddGroupMessageInterceptor(func(m *cqbot.GroupMessage) bool {
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
		return false
	})

	cqbotClient.AddGroupMessageInterceptor(func(m *cqbot.GroupMessage) bool {
		if !strings.Contains(*m.Message, "炮粉") {
			return false
		}
		requestLimitKey := fmt.Sprintf("cqbot:request:limit:%s:%s", strconv.FormatInt(*m.UserId, 10), *m.Message)
		boolCmd := redisClient.SetNX(requestLimitKey, 1, 5*time.Second)
		if !boolCmd.Val() {
			log.Println("redis boolCmd: ", boolCmd)
			content, err := findRandomMessagePhrase("repeat")
			if err != nil {
				panic(err)
			}
			cqbotClient.SendMessage(content, *m.GroupId)
			return true
		}

		return false
	})

	cqbotClient.AddGroupMessageHandler(func(c *cqbot.GroupContext) {
		if !strings.HasPrefix(*c.Message.Message, "炮粉给我干死") {
			return
		}

		r := regexp.MustCompile("炮粉给我干死(.+)")
		keywords := r.FindStringSubmatch(*c.Message.Message)
		if len(keywords) < 2 {
			return
		}
		atUserId, found := findContentByAlias("qq_number", keywords[1])
		if !found {
			log.Println("find alias error: ")
			cqbotClient.SendMessage("少你妈火，没设置alias干死尼玛干死", *c.Message.GroupId)
			return
		}
		if atUserId == strconv.FormatInt(c.LoginInfo.UserId, 10) {
			cqbotClient.SendMessage("你傻逼还是我傻逼？", *c.Message.GroupId)
			return
		}

		words, err := queryRandomMessagePhrase("fuck", 3)
		if err != nil {
			panic(err)
		}
		for _, word := range words {
			cqbotClient.SendMessage(word+buildAt(atUserId), *c.Message.GroupId)
			time.Sleep(2 * time.Second)
		}
	})

	cqbotClient.AddGroupMessageHandler(func(c *cqbot.GroupContext) {
		if *c.Message.Message == "炮粉干我" {
			cqbotClient.SendMessage("真尼玛贱", *c.Message.GroupId)
			words, err := findRandomMessagePhrase("fuck")
			if err != nil {
				panic(err)
			}
			cqbotClient.SendMessage(words, *c.Message.GroupId)
		}
	})

	cqbotClient.AddGroupMessageHandler(func(c *cqbot.GroupContext) {
		if !strings.HasPrefix(*c.Message.Message, "set") && !strings.HasPrefix(*c.Message.Message, "reset") {
			return
		}

		r := regexp.MustCompile("(?P<cmdType>set|reset) (?P<cmd>.+?) (?P<content>.+)")
		submatch := r.FindStringSubmatch(*c.Message.Message)
		cmdMap := make(map[string]string)

		if len(submatch) == len(r.SubexpNames()) {
			// TODO 提取工具包
			for i, name := range r.SubexpNames() {
				if i != 0 && name != "" {
					cmdMap[name] = submatch[i]
				}
			}
		}

		cmd := cmdMap["cmd"]
		content := cmdMap["content"]
		cmdType := cmdMap["cmdType"]
		if content == "" {
			return
		}

		if cmd == "fuck" {
			err := saveMessagePhrase("fuck", content)
			if err != nil {
				log.Println(err)
				cqbotClient.SendMessage("Set failed 并不能阻止我甘玲娘", *c.Message.GroupId)
				return
			}
		} else if cmd == "alias" {
			// set alias 123456789 赵炮
			split := strings.Split(content, " ")
			if len(split) != 2 {
				cqbotClient.SendMessage("Please use [set alias 123456789 customAlias]", *c.Message.GroupId)
				return
			}
			userId := split[0]
			alias := split[1]
			savedAlias, found := findAliasByContent("qq_number", userId)
			if found {
				if cmdType == "reset" {
					updateAlias("qq_number", userId, alias)
					cqbotClient.SendMessage("Set success", *c.Message.GroupId)
					return
				}
				cqbotClient.SendMessage(fmt.Sprintf("尼玛之前就设置别名是%s了", savedAlias), *c.Message.GroupId)
				return
			}
			saveAlias("qq_number", userId, alias)
		} else if cmd == "steam64" {
			split := strings.Split(content, " ")
			if len(split) != 2 {
				cqbotClient.SendMessage("Please use [set steam64 76561198129256636 二狗]", *c.Message.GroupId)
				return
			}
			steam64 := split[0]
			alias := split[1]
			saveAlias("steam64", alias, steam64)
		} else {
			cqbotClient.SendMessage("你set你妈了个蹭次呢？", *c.Message.GroupId)
		}
		cqbotClient.SendMessage("Set success", *c.Message.GroupId)
	})

	r3 := regexp.MustCompile("炮粉查查(.+)最近的比赛")
	cqbotClient.AddGroupMessageHandler(func(c *cqbot.GroupContext) {
		submatch := r3.FindStringSubmatch(*c.Message.Message)
		if len(submatch) != 2 {
			return
		}

		name := submatch[1]
		steam64Id, found := findContentByAlias("steam64", name)
		if !found {
			cqbotClient.SendMessage("没设置steam64查你妈了个崩薄", *c.Message.GroupId)
			return
		}

		uintSteamId, err := strconv.ParseUint(steam64Id, 10, 64)
		if err != nil {
			panic(err)
		}
		recentMatchs := GetRecentMatches(ConvertSteamId64ToSteamAccountId(uintSteamId))

		result := ""
		for _, match := range recentMatchs[:10] {
			kda := (match.Kills + match.Assists) / match.Deaths
			comment := ""
			if kda > 30 {
				comment = fmt.Sprintf("你比样开挂的吧，kda%d？", kda)
			} else if kda > 20 {
				comment = fmt.Sprintf("kda%d？羡慕练小号的", kda)
			} else if kda > 10 {
				comment = fmt.Sprintf("kda%d，HGTV复兴有望好吧", kda)
			} else if kda > 5 {
				comment = fmt.Sprintf("KDA到%d了，不错", kda)
			} else if kda > 3 {
				comment = fmt.Sprintf("KDA%d，马马虎虎", kda)
			} else if kda > 1 {
				comment = fmt.Sprintf("真你妈菜，kda才%d，我代表老板干死你！", kda)
			} else {
				comment = fmt.Sprintf("KDA才%d你还是HGTV的人吗？", kda)
			}

			nowString := time.Unix(int64(match.StartTime), 0).Format("2006-01-02 15:04")
			result = result + fmt.Sprintf("%s打了%d分钟，%s\r\n",
				nowString, match.Duration/60, comment)
		}
		cqbotClient.SendMessage(result, *c.Message.GroupId)
	})

	cqbotClient.Run("0.0.0.0:" + *port)
}

func saveAlias(tp, alias, content string) {
	_, err := db.Exec("insert into cqbot_alias (pk_id, type, alias, value) values (?,?,?,?)", id.Id(), tp, alias, content)
	if err != nil {
		panic(err)
	}
}

func updateAlias(tp, alias, content string) {
	_, err := db.Exec("update cqbot_alias set alias = ? where type = ? and value = ?", alias, tp, content)
	if err != nil {
		panic(err)
	}
}

func findContentByAlias(tp, alias string) (content string, found bool) {
	row := db.QueryRow("select value from cqbot_alias where type = ? and alias = ?", tp, alias)
	err := row.Scan(&content)
	if err != nil {
		return
	}
	return content, true
}

func findAliasByContent(tp, content string) (savedAlias string, found bool) {
	row := db.QueryRow("select alias from cqbot_alias where type = ? and value = ?", tp, content)
	err := row.Scan(&savedAlias)
	if err != nil {
		return
	}
	return savedAlias, true
}

func findRandomMessagePhrase(t string) (content string, err error) {
	row := db.QueryRow("select content from cqbot_message_phrase where type = ? order by rand() limit 1", t)
	err = row.Scan(&content)
	return
}

func queryRandomMessagePhrase(t string, times int) (contents []string, err error) {
	rows, err := db.Query("select content from cqbot_message_phrase where type = ? order by rand() limit ?", t, times)
	if err != nil {
		return
	}

	for rows.Next() {
		content := ""
		err = rows.Scan(&content)
		if err != nil {
			return
		}
		contents = append(contents, content)
	}
	return
}

func saveMessagePhrase(t, words string) error {
	row := db.QueryRow("select content from cqbot_message_phrase where content = ?", words)
	var content string
	err := row.Scan(&content)
	if err == nil || content != "" {
		return errors.New("words already exists")
	}

	_, err = db.Exec("insert into cqbot_message_phrase (pk_id, type, content) values (?, ?,?)", id.Id(), t, words)
	return err
}

func buildAt(userId string) string {
	return fmt.Sprintf("[CQ:at,qq=%s]", userId)
}

func GetRecentMatches(accountId string) (recentMatchs []RecentMatch) {
	recentMatchUrl := "https://api.opendota.com/api/players/{playerId}/recentMatches"
	resp, err := http.Get(strings.Replace(recentMatchUrl, "{playerId}", accountId, -1))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bytes, &recentMatchs)
	if err != nil {
		panic(err)
	}
	return
}

func ConvertSteamId64ToSteamAccountId(steam64Id uint64) string {
	steamId3 := steamid.ParseCommunityID(steam64Id, steamid.AccountTypeIndividual).SteamID3()
	split := strings.Split(steamId3, ":")
	return strings.TrimRight(split[2], "]")
}
