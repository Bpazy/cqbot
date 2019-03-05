package cqbot

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
)

type Config struct {
	HttpApiAddr string
}

type Client struct {
	Config
	// 私聊消息处理器
	privateMessageHandlers []PrivateMessageHandler
	// 群聊消息处理器
	groupMessageHandlers []GroupMessageHandler
	// 讨论组消息处理器
	discussMessageHandlers []DiscussMessageHandler
	// 群消息拦截器
	groupMessageInterceptors []GroupMessageInterceptor
}

func NewClient(httpApiAddr string) *Client {
	return &Client{
		Config: Config{HttpApiAddr: httpApiAddr},
	}
}

func (client *Client) AddPrivateMessageHandler(handler PrivateMessageHandler) {
	client.privateMessageHandlers = append(client.privateMessageHandlers, handler)
}

func (client *Client) AddGroupMessageHandler(handler GroupMessageHandler) {
	client.groupMessageHandlers = append(client.groupMessageHandlers, handler)
}

func (client *Client) AddDiscussMessageHandler(handler DiscussMessageHandler) {
	client.discussMessageHandlers = append(client.discussMessageHandlers, handler)
}

func (client *Client) AddGroupMessageInterceptor(interceptor GroupMessageInterceptor) {
	client.groupMessageInterceptors = append(client.groupMessageInterceptors, interceptor)
}

func (client *Client) Run(addr string) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/", func(c *gin.Context) {
		reply, err := client.dispatchMsg(c)
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

type GroupMessageHandler func(message *GroupMessage)

type DiscussMessageHandler func(message *DiscussMessage)

type GroupMessageInterceptor func(message *GroupMessage) bool

func (client *Client) dispatchMsg(c *gin.Context) (interface{}, error) {
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
			err = client.handlePrivateMessage(bodyBytes)
		}
		if messagePostType.MessageType == group {
			err = client.handleGroupMessage(bodyBytes)
		}
		if messagePostType.MessageType == discuss {
			err = client.handleDiscussMessage(bodyBytes)
		}
	} else {
		log.Printf("%s: %s", postType.PostType, string(bodyBytes))
	}

	if err != nil {
		return nil, err
	}

	return reply, nil
}

func (client *Client) handleDiscussMessage(bytes []byte) error {
	message := DiscussMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", message.MessageType, message.Sender.Nickname, message.UserId, message.Message)

	for _, handler := range client.discussMessageHandlers {
		handler(&message)
	}
	return nil
}

func (client *Client) handleGroupMessage(bytes []byte) error {
	message := GroupMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", *message.MessageType, *message.Sender.Nickname, message.UserId, *message.Message)

	// execute interceptor
	for _, interceptor := range client.groupMessageInterceptors {
		pass := interceptor(&message)
		if pass {
			return nil
		}
	}

	// message handler
	for _, handler := range client.groupMessageHandlers {
		handler(&message)
	}
	return nil
}

func (client *Client) handlePrivateMessage(bytes []byte) error {
	message := PrivateMessage{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("[%s] %s(%d) say: %s", *message.MessageType, *message.Sender.Nickname, message.UserId, *message.Message)

	for _, handler := range client.privateMessageHandlers {
		handler(&message)
	}
	return nil
}

func (client *Client) SendMessage(message string, groupId int64) {
	groupMessageUrl := client.HttpApiAddr + "/send_group_msg"
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
