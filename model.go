package main

type KeywordInfo struct {
	UserId   string `db:"user_id"`
	Nickname string `db:"nickname"`
	Count    int    `db:"count"`
}

type RecentMatch struct {
	MatchID      int64       `json:"match_id"`
	PlayerSlot   int         `json:"player_slot"`
	RadiantWin   bool        `json:"radiant_win"`
	Duration     int         `json:"duration"`
	GameMode     int         `json:"game_mode"`
	LobbyType    int         `json:"lobby_type"`
	HeroID       int         `json:"hero_id"`
	StartTime    int         `json:"start_time"`
	Version      interface{} `json:"version"`
	Kills        int         `json:"kills"`
	Deaths       int         `json:"deaths"`
	Assists      int         `json:"assists"`
	Skill        int         `json:"skill"`
	XpPerMin     int         `json:"xp_per_min"`
	GoldPerMin   int         `json:"gold_per_min"`
	HeroDamage   int         `json:"hero_damage"`
	TowerDamage  int         `json:"tower_damage"`
	HeroHealing  int         `json:"hero_healing"`
	LastHits     int         `json:"last_hits"`
	Lane         interface{} `json:"lane"`
	LaneRole     interface{} `json:"lane_role"`
	IsRoaming    interface{} `json:"is_roaming"`
	Cluster      int         `json:"cluster"`
	LeaverStatus int         `json:"leaver_status"`
	PartySize    interface{} `json:"party_size"`
}
