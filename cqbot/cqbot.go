package cqbot

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
)

func Run(addr string) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/", func(c *gin.Context) {
		err := dispatchMsg(c)
		if err != nil {
			panic(err)
		}
		c.JSON(200, nil)
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

	log.Printf("[%s] %s(%d) say: %s", message.MessageType, message.Sender.Nickname, message.UserId, message.Message)

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

	log.Printf("[%s] %s(%d) say: %s", message.MessageType, message.Sender.Nickname, message.UserId, message.Message)

	for _, handler := range privateMessageHandlers {
		handler(&message)
	}
	return nil
}
