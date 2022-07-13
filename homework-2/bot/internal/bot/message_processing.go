package bot

import (
	"fmt"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/model"
	"strconv"
	"strings"
)

var mainMenuText = fmt.Sprintf("%s\n%s\n%s\n", "/portfolio", "/dashboard", "/close")
var mainMenu = []string{"/portfolio", "/dashboard", "/close"}
var portfolioMenu = fmt.Sprintf("%s\n%s\n%s\n%s\n", "/show", "/new", "/update", "/close")

func ProcessMessages(incoming chan model.InputMessage, outgoing chan model.OutgoingMessage, store SessionStorage, service PortfolioService) {
	for msg := range incoming {
		sessionKey := fmt.Sprintf("%s-%d", msg.Messenger, msg.UserId)
		var session *Session
		session, ok := store.GetSession(sessionKey)
		if !ok {
			session = new(Session)
			session.state = initState
			session.userId = strconv.FormatInt(msg.UserId, 10)
			session.chatId = msg.ChatId
			session.messenger = msg.Messenger
			session.vars = make(map[string]string)
		}
		session.msg = msg.Text
		account, err := service.GetOrCreateAccount(session.messenger, session.userId)
		if err != nil {
			outgoing <- model.TextMessage(session.messenger, session.chatId, "unexpected error")
			session.state = initState
		}
		session.accountId = account.Id
		store.PutSession(sessionKey, session)
		session.state(session, outgoing, service)
	}
}

func initState(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) {
	handleMenu(session, outgoing, service)
}

func showDashboard(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) {
	dashboard, err := service.GetDashboard(session.accountId)
	if err != nil {
		session.state = initState
		outgoing <- model.TextMessage(session.messenger, session.chatId, "Unable to show dashboard")
		outgoing <- model.TextMessage(session.messenger, session.chatId, mainMenuText)
	}
	var msgB strings.Builder
	for _, p := range dashboard.ValueList {
		msgB.WriteString(fmt.Sprintf("%s — %.2f\n", p.Name, p.Value))
	}
	msg := fmt.Sprintf("Dashboard:\nCurrent account value: %.2f\n%s\n", dashboard.TotalValue, msgB.String())
	outgoing <- model.TextMessage(session.messenger, session.chatId, msg)
}

func portfolioState(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) {
	if handleMenu(session, outgoing, service) {
		return
	}
	switch session.msg {
	case "/show":
		portfolioList, err := service.GetPortfolioList(session.accountId)
		if err != nil {
			msgText := fmt.Sprint("fail to get portfolio list")
			outgoing <- model.TextMessage(session.messenger, session.chatId, msgText)
			outgoing <- model.TextMessage(session.messenger, session.chatId, portfolioMenu)
			session.state = portfolioState
			return
		}
		if len(portfolioList) == 0 {
			outgoing <- model.TextMessage(session.messenger, session.chatId, "No created portfolio")
			session.state = portfolioState
		} else {
			options := make([]string, 0, len(portfolioList))
			for _, p := range portfolioList {
				options = append(options, p.Name)
			}
			outgoing <- model.OptionsMessage(session.messenger, session.chatId, options)
			session.state = portfolioShowWaitName
		}
	case "/new":
		session.state = portfolioNewWaitName
		msgText := fmt.Sprintf("%s\n", "Enter portfolio name:")
		outgoing <- model.TextMessage(session.messenger, session.chatId, msgText)
	case "/update":
		portfolioList, err := service.GetPortfolioList(session.accountId)
		if err != nil {
			msgText := fmt.Sprint("fail to get portfolio list")
			outgoing <- model.TextMessage(session.messenger, session.chatId, msgText)
			outgoing <- model.TextMessage(session.messenger, session.chatId, portfolioMenu)
			session.state = portfolioState
			return
		}
		if len(portfolioList) == 0 {
			session.state = portfolioState
			outgoing <- model.TextMessage(session.messenger, session.chatId, "No created portfolio")
		} else {
			options := make([]string, 0, len(portfolioList))
			for _, p := range portfolioList {
				options = append(options, p.Name)
			}
			session.state = portfolioUpdateWaitName
			outgoing <- model.TextMessage(session.messenger, session.chatId, "Portfolio:")
			outgoing <- model.OptionsMessage(session.messenger, session.chatId, options)
		}
	case "/close":
		session.state = initState
		outgoing <- model.TextMessage(session.messenger, session.chatId, mainMenuText)
	}
}

func portfolioShowWaitName(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) {
	if handleMenu(session, outgoing, service) {
		return
	}
	portfolioName := session.msg

	portfolio, err := service.GetPortfolio(session.accountId, portfolioName)
	if err != nil {
		session.state = portfolioState
		outgoing <- model.TextMessage(session.messenger, session.chatId, err.Error())
		outgoing <- model.TextMessage(session.messenger, session.chatId, portfolioMenu)
		return
	}
	msg := strings.Builder{}
	msg.WriteString(fmt.Sprintf("%s\t%s\n", "Symbol", "Qty"))
	for _, pp := range portfolio.Positions {
		msg.WriteString(fmt.Sprintf("%s\t%d\n", pp.Symbol, pp.Quantity))
	}

	outgoing <- model.TextMessage(session.messenger, session.chatId, msg.String())

	session.state = portfolioState
	outgoing <- model.TextMessage(session.messenger, session.chatId, portfolioMenu)
}

func portfolioUpdateWaitName(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) {
	if handleMenu(session, outgoing, service) {
		return
	}
	portfolioName := session.msg
	portfolio, err := service.GetPortfolio(session.accountId, portfolioName)
	if err != nil {
		session.state = portfolioState
		outgoing <- model.TextMessage(session.messenger, session.chatId, err.Error())
		outgoing <- model.TextMessage(session.messenger, session.chatId, portfolioMenu)
		return
	}
	session.vars["portfolioId"] = strconv.FormatInt(int64(portfolio.Id), 10)
	session.state = portfolioUpdateWaitSymbol
	msgText := fmt.Sprintf("Enter market symbol\nex. OZON")
	outgoing <- model.TextMessage(session.messenger, session.chatId, msgText)
}

func portfolioUpdateWaitSymbol(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) {
	if handleMenu(session, outgoing, service) {
		return
	}
	symbol := session.msg
	session.vars["symbol"] = symbol
	session.state = portfolioUpdateWaitQyt
	msgText := fmt.Sprintf("Enter quantity with space:\nex. 42")
	outgoing <- model.TextMessage(session.messenger, session.chatId, msgText)
}

func portfolioUpdateWaitQyt(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) {
	if handleMenu(session, outgoing, service) {
		return
	}
	quantity, err := strconv.ParseInt(session.msg, 10, 32)
	if err != nil {
		session.state = portfolioState
		outgoing <- model.TextMessage(session.messenger, session.chatId, "Invalid number")
	}
	portfolioId, _ := strconv.ParseInt(session.vars["portfolioId"], 10, 32)
	symbol := session.vars["symbol"]
	err = service.NewPosition(int32(portfolioId), symbol, int32(quantity))
	if err != nil {
		outgoing <- model.TextMessage(session.messenger, session.chatId, "failed to create portfolio position: "+err.Error())
	}

	session.state = portfolioState
	outgoing <- model.TextMessage(session.messenger, session.chatId, portfolioMenu)
}

func portfolioNewWaitName(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) {
	if handleMenu(session, outgoing, service) {
		return
	}
	portfolioName := session.msg
	portfolio, err := service.CreatePortfolio(session.accountId, portfolioName)
	if err != nil {
		session.state = portfolioState
		outgoing <- model.TextMessage(session.messenger, session.chatId, err.Error())
		outgoing <- model.TextMessage(session.messenger, session.chatId, mainMenuText)
		return

	}
	session.state = portfolioState
	msgText := fmt.Sprintf("New portfolio:\n%s", portfolio.Name)
	outgoing <- model.TextMessage(session.messenger, session.chatId, msgText)
	outgoing <- model.TextMessage(session.messenger, session.chatId, portfolioMenu)
}

func handleMenu(session *Session, outgoing chan model.OutgoingMessage, service PortfolioService) bool {
	switch session.msg {
	case "/start":
		outgoing <- model.MenuMessage(session.messenger, session.chatId, mainMenu)
		return true
	case "/portfolio":
		session.state = portfolioState
		outgoing <- model.TextMessage(session.messenger, session.chatId, portfolioMenu)
		return true
	case "/dashboard":
		showDashboard(session, outgoing, service)
		session.state = initState
		outgoing <- model.TextMessage(session.messenger, session.chatId, mainMenuText)
		return true
	case "/close":
		session.state = initState
		outgoing <- model.TextMessage(session.messenger, session.chatId, mainMenuText)
		return true
	}
	return false
}
