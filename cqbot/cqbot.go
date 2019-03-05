package cqbot

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
)

func Run(addr string) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/", func(c *gin.Context) {
		reply, err := dispatchMsg(c)
		if err != nil {
			panic(err)
		}
		c.JSON(200, reply)
	})
	_ = r.Run(addr)
}

// post type
const (
	private = "private"
	group   = "group"
	discuss = "discuss"
)

type PrivateMessageHandler func(message *PrivateMessage)

var privateMessageHandlers []PrivateMessageHandler

type GroupMessageHandler func(message *GroupMessage)

var groupMessageHandlers []GroupMessageHandler

type DiscussMessageHandler func(message *DiscussMessage)

var discussMessageHandlers []DiscussMessageHandler

func AddPrivateMessageHandler(handler PrivateMessageHandler) {
	privateMessageHandlers = append(privateMessageHandlers, handler)
}
func AddGroupMessageHandler(handler GroupMessageHandler) {
	groupMessageHandlers = append(groupMessageHandlers, handler)
}
func AddDiscussMessageHandler(handler DiscussMessageHandler) {
	discussMessageHandlers = append(discussMessageHandlers, handler)
}

// return true if stop pass
type GroupMessageInterceptor func(message *GroupMessage) bool

var groupMessageInterceptors []GroupMessageInterceptor

func AddGroupMessageInterceptor(interceptor GroupMessageInterceptor) {
	groupMessageInterceptors = append(groupMessageInterceptors, interceptor)
}

func dispatchMsg(c *gin.Context) (interface{}, error) {
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	defer c.Request.Body.Close()

	postType := PostType{}
	err = json.Unmarshal(bodyBytes, &postType)
	if err != nil {
		return nil, err
	}

	var reply interface{}
	if postType.PostType == "message" {
		messagePostType := MessagePostType{}
		err = json.Unmarshal(bodyBytes, &messagePostType)
		if err != nil {
			return nil, err
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
		return nil, err
	}

	return reply, nil
}

func handleDiscussMessage(bytes []byte) error {
	message := DiscussMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", message.MessageType, message.Sender.Nickname, message.UserId, message.Message)

	for _, handler := range discussMessageHandlers {
		handler(&message)
	}
	return nil
}

func handleGroupMessage(bytes []byte) error {
	message := GroupMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", *message.MessageType, *message.Sender.Nickname, message.UserId, *message.Message)

	// execute interceptor
	for _, interceptor := range groupMessageInterceptors {
		pass := interceptor(&message)
		if pass {
			break
		}
	}

	// message handler
	for _, handler := range groupMessageHandlers {
		handler(&message)
	}
	return nil
}

func handlePrivateMessage(bytes []byte) error {
	message := PrivateMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", *message.MessageType, *message.Sender.Nickname, message.UserId, *message.Message)

	for _, handler := range privateMessageHandlers {
		handler(&message)
	}
	return nil
}

func SendMessage(message string, groupId int64) {
	groupMessageUrl := "http://127.0.0.1:5701/send_group_msg"
	m := map[string]interface{}{
		"message":     message,
		"group_id":    groupId,
		"auto_escape": false,
	}

	jsonStr, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post(groupMessageUrl, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("api response Body:", string(body))
}
