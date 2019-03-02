package main

type Xunsu struct {
	UserId   string `db:"user_id"`
	Nickname string `db:"nickname"`
	Count    int    `db:"count"`
}
