package model

import "strings"

type Messenger string

const (
	Telegram Messenger = "Telegram"
)

func (m Messenger) String() string {
	return strings.ToLower(string(m))
}

func NewMessenger(m string) (Messenger, bool) {
	switch m {
	case string(Telegram):
		return Telegram, true
	}
	return "", false
}

type Account struct {
	Id          int32     `db:"id"`
	Messenger   Messenger `db:"messenger"`
	MessengerId string    `db:"messenger_id"`
}

type CreateAccountReq struct {
	Messenger   Messenger
	MessengerId string
}
