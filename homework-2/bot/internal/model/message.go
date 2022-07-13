package model

type MessageType int32

const (
	Text    MessageType = 0
	Menu    MessageType = 1
	Options MessageType = 2
)

type InputMessage struct {
	Messenger string
	UserId    int64
	ChatId    int64
	Text      string
}

type OutgoingMessage struct {
	Messenger string
	ChatId    int64
	Values    []string
	Type      MessageType
}

func TextMessage(messenger string, chatId int64, textMsg string) OutgoingMessage {
	return OutgoingMessage{
		Messenger: messenger,
		ChatId:    chatId,
		Values:    []string{textMsg},
		Type:      Text,
	}
}

func OptionsMessage(messenger string, chatId int64, options []string) OutgoingMessage {
	return OutgoingMessage{
		Messenger: messenger,
		ChatId:    chatId,
		Values:    options,
		Type:      Options,
	}
}

func MenuMessage(messenger string, chatId int64, menu []string) OutgoingMessage {
	return OutgoingMessage{
		Messenger: messenger,
		ChatId:    chatId,
		Values:    menu,
		Type:      Menu,
	}
}
