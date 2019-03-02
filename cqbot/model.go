package cqbot

import "time"

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
