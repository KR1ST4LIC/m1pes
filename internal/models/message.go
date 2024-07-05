package models

type Message struct {
	User   User
	Coin   Coin
	Action string
	File   string
	Line   int
}

type Error struct {
	File string
	Line int
}
