package bot

import (
	"context"
	"errors"
	"fmt"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/logging"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/bot/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type PortfolioService interface {
	CreatePortfolio(int32, string) (*model.Portfolio, error)
	GetPortfolioList(int32) ([]model.Portfolio, error)
	GetPortfolio(int32, string) (*model.Portfolio, error)
	NewPosition(int32, string, int32) error
	GetDashboard(int32) (*model.Dashboard, error)
	GetOrCreateAccount(string, string) (*model.Account, error)
	CreateAccount(string, string) (*model.Account, error)
}

type DefaultPortfolioService struct {
	client pb.PortfolioServiceClient
}

func NewDefaultPortfolioService(client pb.PortfolioServiceClient) DefaultPortfolioService {
	return DefaultPortfolioService{client: client}
}

const RequestTimeout = 10 * time.Second

func (srv DefaultPortfolioService) CreatePortfolio(accountId int32, portfolioName string) (*model.Portfolio, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := pb.CreatePortfolioRequest{
		AccountId: accountId,
		Name:      portfolioName,
	}
	portfolio, err := srv.client.CreatePortfolio(ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.InvalidArgument {
				return nil, errors.New(e.Message())
			}
		} else {
			logging.Error(fmt.Sprintf("Unable to parse error from grcp: %v", err))
			return nil, fmt.Errorf("failed to create new portfolio")
		}
		logging.Error(fmt.Sprintf("Fail to create portfolio: " + err.Error()))
		return nil, fmt.Errorf("failed to create new portfolio")
	}

	return &model.Portfolio{Id: portfolio.Portfolio.Id, Name: portfolio.Portfolio.Name}, nil
}

func (srv DefaultPortfolioService) GetOrCreateAccount(messenger string, messengerId string) (*model.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	req := pb.GetAccountInfoRequest{
		MessengerId: messengerId,
		Messenger:   pb.Messenger(pb.Messenger_value[messenger]),
	}
	acc, err := srv.client.GetAccount(ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return srv.CreateAccount(messenger, messengerId)
			}
		} else {
			logging.Error(fmt.Sprintf("Unable to parse error from grcp: %v", err))
			return nil, errors.New("fail to get account")
		}
	}

	return &model.Account{Id: acc.GetAccount().GetId()}, nil
}

func (srv DefaultPortfolioService) CreateAccount(messenger string, messengerId string) (*model.Account, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	messengerPb := pb.Messenger(pb.Messenger_value[messenger])
	createReq := pb.CreateAccountRequest{
		Messenger:   messengerPb,
		MessengerId: messengerId,
	}
	newAcc, err := srv.client.CreateAccount(ctx, &createReq)
	if err != nil {
		logging.Error("Failed to create account: " + err.Error())
		return nil, errors.New("fail to create account")
	}
	return &model.Account{Id: newAcc.GetAccount().GetId()}, err
}

func (srv DefaultPortfolioService) GetPortfolioList(accountId int32) ([]model.Portfolio, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	portfolioReq := pb.GetAccountPortfolioRequest{AccountId: accountId}
	portfolioResp, err := srv.client.GetAccountPortfolio(ctx, &portfolioReq)
	if err != nil {
		logging.Error("Error getting account's portfolio list: " + err.Error())
		return nil, errors.New("fail to get portfolio list")
	}
	portfolioList := make([]model.Portfolio, 0, len(portfolioResp.PortfolioList))
	for _, p := range portfolioResp.PortfolioList {
		positions := make([]model.Position, 0)
		for _, pp := range p.Positions {
			position := model.Position{
				Id:            pp.GetId(),
				Symbol:        pp.GetAssetCode(),
				Quantity:      pp.GetQuantity(),
				PlacementTime: pp.PlacementTime.AsTime(),
			}
			positions = append(positions, position)
		}

		portfolioList = append(portfolioList, model.Portfolio{Id: p.Id, Name: p.Name, Positions: positions})
	}
	return portfolioList, nil
}

func (srv DefaultPortfolioService) NewPosition(portfolioId int32, symbol string, quantity int32) error {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	positionReq := pb.NewPortfolioPositionRequest{
		PortfolioId: portfolioId,
		AssetCode:   symbol,
		Quantity:    quantity,
	}
	_, err := srv.client.NewPortfolioPosition(ctx, &positionReq)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.InvalidArgument {
				return errors.New(e.Message())
			}
		} else {
			logging.Error(fmt.Sprintf("Unable to parse error from grcp: %v", err))
			return errors.New("unexpected error")
		}
		logging.Error("Failed to create portfolio position: " + err.Error())
		return errors.New("fail to create portfolio position")
	}

	return nil
}

func (srv DefaultPortfolioService) GetPortfolio(accountId int32, portfolioName string) (*model.Portfolio, error) {
	list, err := srv.GetPortfolioList(accountId)
	if err != nil {
		return nil, err
	}
	var portfolio *model.Portfolio
	for _, p := range list {
		if p.Name == portfolioName {
			portfolio = &p
		}
	}
	if portfolio == nil {
		return nil, errors.New(fmt.Sprintf("not found portfolio '%s'", portfolioName))
	}
	return portfolio, nil
}

func (srv DefaultPortfolioService) GetDashboard(accountId int32) (*model.Dashboard, error) {
	ctx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	req := pb.GetAccountDashboardRequest{AccountId: accountId}
	pbDashboard, err := srv.client.GetAccountDashboard(ctx, &req)
	if err != nil {
		logging.Error("Error getting dashboard: " + err.Error())
		return nil, errors.New(fmt.Sprintf("fail to get dashboard"))
	}

	valueList := make([]model.PortfolioValue, 0)
	for _, pbValue := range pbDashboard.GetDashboard().PortfolioValueList {
		valueList = append(valueList, model.PortfolioValue{
			Id:    pbValue.PortfolioId,
			Name:  pbValue.PortfolioName,
			Value: pbValue.Value,
		})
	}
	dashboard := model.Dashboard{
		TotalValue: pbDashboard.GetDashboard().TotalValue,
		ValueList:  valueList,
	}
	return &dashboard, nil
}
