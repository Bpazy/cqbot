package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
)

func main() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/", func(c *gin.Context) {
		err := dispatchMsg(c)
		if err != nil {
			log.Println(err)
			c.Abort()
		}
		c.JSON(200, nil)
	})
	_ = r.Run("0.0.0.0:12345") // listen and serve on 0.0.0.0:8080
}

type PostType struct {
	PostType string `json:"post_type"`
}

type Message struct {
	PostType    string `json:"post_type"`    // possible value: message
	MessageType string `json:"message_type"` // possible value: private
	SubType     string `json:"sub_type"`     // possible value: friend,group,discuss,other
	MessageId   int32  `json:"message_id"`
	UserId      int64  `json:"user_id"`
	Message     string `json:"message"`
	RawMessage  string `json:"raw_message"`
	Font        int32  `json:"font"`
	Sender      Sender `json:"sender"`
}

type Sender struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Sex      string `json:"sex"`
	Age      int32  `json:"age"`
}

// TODO feature
type Notice struct {
}

// TODO feature
type Request struct {
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
		err = handleMessage(bodyBytes)
	}

	if err != nil {
		return err
	}

	return nil
}

func handleMessage(bytes []byte) error {
	message := Message{}
	err := json.Unmarshal(bytes, &message)
	if err != nil {
		return err
	}

	log.Printf("%+v", message)
	return nil
}
