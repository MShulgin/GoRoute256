package server

import (
	"context"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/pb"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/service"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedPortfolioServiceServer
	accountService   service.AccountService
	portfolioService service.PortfolioService
	dashboardService service.DashboardService
}

func NewServer(accService service.AccountService, portfolioService service.PortfolioService,
	dashboardService service.DashboardService) Server {
	server := Server{
		accountService:   accService,
		portfolioService: portfolioService,
		dashboardService: dashboardService,
	}
	return server
}

func (s Server) GetAccount(_ context.Context, request *pb.GetAccountInfoRequest) (*pb.AccountInfoResponse, error) {
	messenger, _ := model.NewMessenger(request.GetMessenger().String())
	a, err := s.accountService.GetAccount(messenger, request.GetMessengerId())
	if err != nil {
		err := status.Errorf(err.GrpcCode(), err.Message)
		return nil, err
	}
	msg := pb.Messenger_value[string(a.Messenger)]
	accountDto := pb.Account{Id: a.Id, Messenger: pb.Messenger(msg)}
	return &pb.AccountInfoResponse{Account: &accountDto}, nil
}

func (s Server) CreateAccount(_ context.Context, request *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	newAccReq := model.CreateAccountReq{
		Messenger:   model.Messenger(request.GetMessenger().String()),
		MessengerId: request.GetMessengerId(),
	}
	a, err := s.accountService.NewAccount(newAccReq)
	if err != nil {
		err := status.Errorf(err.GrpcCode(), err.Message)
		return nil, err
	}
	r := pb.Account{Id: a.Id, Messenger: pb.Messenger(pb.Messenger_value[string(a.Messenger)])}
	return &pb.CreateAccountResponse{Account: &r}, nil
}

func (s Server) CreatePortfolio(_ context.Context, request *pb.CreatePortfolioRequest) (*pb.CreatePortfolioResponse, error) {
	portfolioReq := model.NewPortfolioRequest{
		AccountId: request.GetAccountId(),
		Name:      request.GetName(),
	}
	p, err := s.portfolioService.NewPortfolio(portfolioReq)
	if err != nil {
		err := status.Errorf(err.GrpcCode(), err.Message)
		return nil, err
	}
	resp := pb.Portfolio{
		Id:   p.Id,
		Name: p.Name,
	}
	return &pb.CreatePortfolioResponse{Portfolio: &resp}, nil
}

func (s Server) GetAccountPortfolio(_ context.Context, request *pb.GetAccountPortfolioRequest) (*pb.GetPortfolioResponse, error) {
	accountId := request.GetAccountId()
	_, err := s.accountService.GetAccountById(accountId)
	if err != nil {
		err := status.Errorf(err.GrpcCode(), err.Message)
		return nil, err
	}
	portfolioList, err := s.portfolioService.GetAccountPortfolio(accountId)
	if err != nil {
		err := status.Errorf(err.GrpcCode(), err.Message)
		return nil, err
	}

	responseList := make([]*pb.Portfolio, 0)
	for _, p := range portfolioList {
		pbPositions := make([]*pb.PortfolioPosition, 0)
		for _, pp := range p.Positions {
			pbPositions = append(pbPositions, &pb.PortfolioPosition{
				Id:            pp.Id,
				AssetCode:     pp.Symbol,
				Quantity:      pp.Quantity,
				PlacementTime: timestamppb.New(pp.PlacementTime),
			})
		}
		responseList = append(responseList, &pb.Portfolio{
			Id:        p.Id,
			Name:      p.Name,
			Positions: pbPositions,
		})
	}

	return &pb.GetPortfolioResponse{PortfolioList: responseList}, nil
}

func (s Server) NewPortfolioPosition(_ context.Context, request *pb.NewPortfolioPositionRequest) (*pb.NewPortfolioPositionResponse, error) {
	portfolioId := request.GetPortfolioId()
	newPositionReq := model.NewPositionRequest{
		PortfolioId: portfolioId,
		Symbol:      request.GetAssetCode(),
		Quantity:    request.GetQuantity(),
	}
	newPosition, err := s.portfolioService.PlacePosition(newPositionReq)
	if err != nil {
		err := status.Errorf(err.GrpcCode(), err.Message)
		return nil, err
	}

	pbPosition := pb.PortfolioPosition{
		Id:            newPosition.Id,
		AssetCode:     newPosition.Symbol,
		Quantity:      newPosition.Quantity,
		PlacementTime: timestamppb.New(newPosition.PlacementTime),
	}
	return &pb.NewPortfolioPositionResponse{Position: &pbPosition}, nil
}

func (s Server) DeletePortfolioPosition(_ context.Context, request *pb.DeletePortfolioPositionRequest) (*pb.DeletePortfolioPositionResponse, error) {
	err := s.portfolioService.DeletePosition(request.GetPortfolioId(), request.GetPositionId())
	if err != nil {
		err := status.Errorf(err.GrpcCode(), err.Message)
		return nil, err
	}

	return &pb.DeletePortfolioPositionResponse{}, nil
}

func (s Server) GetAccountDashboard(_ context.Context, request *pb.GetAccountDashboardRequest) (*pb.GetAccountDashboardResponse, error) {
	accountId := request.GetAccountId()
	dashboard, err := s.dashboardService.GetAccountDashboard(accountId)
	if err != nil {
		err := status.Errorf(err.GrpcCode(), err.Message)
		return nil, err
	}

	pbValueList := make([]*pb.PortfolioValueInfo, 0)
	for _, pv := range dashboard.PortfolioValueList {
		pbValueList = append(pbValueList, &pb.PortfolioValueInfo{
			PortfolioId:   pv.PortfolioId,
			PortfolioName: pv.PortfolioName,
			Value:         pv.Value,
		})
	}
	pbDashboard := pb.AccountDashboard{
		AccountId:          dashboard.AccountId,
		TotalValue:         dashboard.TotalValue,
		PortfolioValueList: pbValueList,
	}
	pbResponse := pb.GetAccountDashboardResponse{
		Dashboard: &pbDashboard,
	}
	return &pbResponse, nil
}
