package id

import "github.com/rs/xid"

func Id() string {
	guid := xid.New()
	return guid.String()
}

func PId() *string {
	id := Id()
	return &id
}
