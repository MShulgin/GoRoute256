package main

import (
	"flag"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/bot"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/config"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/logging"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/pb"
	confReader "gitlab.ozon.dev/MShulgin/homework-2/commons/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io/ioutil"
	"strings"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "conf", "config/portfolio.yml", "conf file path")
	flag.Parse()

	conf, err := readConfig(configPath, confReader.NewYamlReader())
	if err != nil {
		panic(err)
	}

	run(conf)
}

func readConfig(configPath string, cfgReader confReader.Reader) (*config.Config, error) {
	confContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		logging.Error("fail to read config file: " + err.Error())
		return nil, err
	}
	var conf config.Config
	err = cfgReader.ReadConfig(confContent, &conf)
	if err != nil {
		logging.Error("fail to parse config: " + err.Error())
		return nil, err
	}
	return &conf, nil
}

func run(conf *config.Config) {
	tgbot, err := tgbotapi.NewBotAPI(conf.Telegram.Token)
	if err != nil {
		logging.Error("fail to create bot api: " + err.Error())
		panic(err)
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(conf.Portfolio.Addr, opts...)
	if err != nil {
		logging.Error("fail to connect to grpc server: " + err.Error())
		panic(err)
	}
	defer conn.Close()

	client := pb.NewPortfolioServiceClient(conn)
	srv := bot.NewDefaultPortfolioService(client)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10
	incomingMessages := make(chan model.InputMessage)
	outgoingMessages := make(chan model.OutgoingMessage)
	sessionStorage := bot.NewMapSessionStorage()

	go bot.ProcessMessages(incomingMessages, outgoingMessages, &sessionStorage, srv)

	go processOutgoingMsg(outgoingMessages, tgbot)

	processTelegramMessages(tgbot.GetUpdatesChan(u), incomingMessages)
}

func processTelegramMessages(updates tgbotapi.UpdatesChannel, incomingMessages chan model.InputMessage) {
	for update := range updates {
		if update.Message != nil {
			tgMsg := update.Message
			srvMsg := model.InputMessage{
				Messenger: "Telegram",
				UserId:    tgMsg.From.ID,
				ChatId:    tgMsg.Chat.ID,
				Text:      tgMsg.Text,
			}
			incomingMessages <- srvMsg
		}
	}
}

func processOutgoingMsg(outgoingMessages chan model.OutgoingMessage, tgbot *tgbotapi.BotAPI) {
	for msg := range outgoingMessages {
		switch msg.Messenger {
		case "Telegram":
			tgMsg := convertToTelegramMessage(msg)
			if _, err := tgbot.Send(tgMsg); err != nil {
				logging.Error("failed to send message: " + err.Error())
			}
		}
	}
}

func convertToTelegramMessage(msg model.OutgoingMessage) tgbotapi.Chattable {
	var m tgbotapi.MessageConfig
	switch msg.Type {
	case model.Text:
		textMsg := strings.Join(msg.Values, "\n")
		m = tgbotapi.NewMessage(msg.ChatId, textMsg)
		m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	case model.Options:
		m = tgbotapi.NewMessage(msg.ChatId, "Options")
		keyboardMarkup := tgbotapi.NewReplyKeyboard()
		for _, v := range msg.Values {
			keyboardRow := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(v)}
			keyboardMarkup.Keyboard = append(keyboardMarkup.Keyboard, keyboardRow)
		}
		m.ReplyMarkup = keyboardMarkup
	}
	return m

}
