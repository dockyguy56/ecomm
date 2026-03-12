package server

import (
	"context"

	"github.com/dockyguy56/ecomm/internal/ecomm-grpc/pb"
	"github.com/dockyguy56/ecomm/internal/ecomm-grpc/storer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	storer *storer.PostgresStorer
	pb.UnimplementedEcommServer
}

func NewServer(storer *storer.PostgresStorer) *Server {
	return &Server{storer: storer}
}

func (s *Server) CreateProduct(ctx context.Context, req *pb.ProductReq) (*pb.ProductRes, error) {
	pr, err := s.storer.CreateProduct(ctx, toStorerProduct(req))
	if err != nil {
		return nil, err
	}

	return toPBProductRes(pr), nil
}

func (s *Server) GetProduct(ctx context.Context, p *pb.ProductReq) (*pb.ProductRes, error) {
	pr, err := s.storer.GetProductByID(ctx, p.GetId())
	if err != nil {
		return nil, err
	}

	return toPBProductRes(pr), nil
}

func (s *Server) GetAllProducts(ctx context.Context, p *pb.ProductReq) (*pb.ListProductRes, error) {
	lps, err := s.storer.GetAllProducts(ctx)
	if err != nil {
		return nil, err
	}

	lpr := make([]*pb.ProductRes, 0, len(lps))
	for _, lp := range lps {
		lpr = append(lpr, toPBProductRes(lp))
	}

	return &pb.ListProductRes{
		Products: lpr,
	}, nil
}

func (s *Server) UpdateProduct(ctx context.Context, p *pb.ProductReq) (*pb.ProductRes, error) {
	product, err := s.storer.GetProductByID(ctx, p.GetId())
	if err != nil {
		return nil, err
	}

	patchProductReq(product, p)
	pr, err := s.storer.UpdateProduct(ctx, product)
	if err != nil {
		return nil, err
	}

	return toPBProductRes(pr), nil
}

func (s *Server) DeleteProduct(ctx context.Context, p *pb.ProductReq) (*pb.ProductRes, error) {
	err := s.storer.DeleteProduct(ctx, p.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.ProductRes{}, nil
}

func (s *Server) CreateOrder(ctx context.Context, o *pb.OrderReq) (*pb.OrderRes, error) {
	order, err := s.storer.CreateOrder(ctx, toStorerOrder(o))
	if err != nil {
		return nil, err
	}

	return toPBOrderRes(order), nil
}

func (s *Server) GetAllOrdersByID(ctx context.Context, o *pb.OrderReq) (*pb.ListOrderRes, error) {
	orders, err := s.storer.GetAllOrdersByID(ctx, o.GetUserId())
	if err != nil {
		return nil, err
	}

	lor := make([]*pb.OrderRes, 0, len(orders))
	for _, lo := range orders {
		lor = append(lor, toPBOrderRes(lo))
	}

	return &pb.ListOrderRes{
		Orders: lor,
	}, nil
}

func (s *Server) GetAllOrders(ctx context.Context, o *pb.OrderReq) (*pb.ListOrderRes, error) {
	orders, err := s.storer.GetAllOrders(ctx)
	if err != nil {
		return nil, err
	}

	lor := make([]*pb.OrderRes, 0, len(orders))
	for _, order := range orders {
		lor = append(lor, toPBOrderRes(order))
	}

	return &pb.ListOrderRes{
		Orders: lor,
	}, nil
}

func (s *Server) DeleteOrder(ctx context.Context, o *pb.OrderReq) (*pb.OrderRes, error) {
	err := s.storer.DeleteOrder(ctx, o.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.OrderRes{}, nil
}

func (s *Server) CreateUser(ctx context.Context, u *pb.UserReq) (*pb.UserRes, error) {
	user, err := s.storer.CreateUser(ctx, toStorerUser(u))
	if err != nil {
		return nil, err
	}

	return toPBUserRes(user), nil
}

func (s *Server) GetUser(ctx context.Context, u *pb.UserReq) (*pb.UserRes, error) {
	user, err := s.storer.GetUser(ctx, u.GetEmail())
	if err != nil {
		return nil, err
	}

	return toPBUserRes(user), nil
}

func (s *Server) GetAllUsers(ctx context.Context, u *pb.UserReq) (*pb.ListUserRes, error) {
	users, err := s.storer.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	lur := make([]*pb.UserRes, 0, len(users))
	for _, user := range users {
		lur = append(lur, toPBUserRes(user))
	}

	return &pb.ListUserRes{
		Users: lur,
	}, nil
}

func (s *Server) UpdateUser(ctx context.Context, u *pb.UserReq) (*pb.UserRes, error) {
	user, err := s.storer.GetUser(ctx, u.GetEmail())
	if err != nil {
		return nil, err
	}

	patchUserReq(user, u)
	ur, err := s.storer.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return toPBUserRes(ur), nil
}

func (s *Server) DeleteUser(ctx context.Context, u *pb.UserReq) (*pb.UserRes, error) {
	err := s.storer.DeleteUser(ctx, u.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.UserRes{}, nil
}

func (s *Server) CreateSession(ctx context.Context, sr *pb.SessionReq) (*pb.SessionRes, error) {
	sess, err := s.storer.CreateSession(ctx, &storer.Session{
		ID:           sr.GetId(),
		UserEmail:    sr.GetUserEmail(),
		RefreshToken: sr.GetRefreshToken(),
		IsRevoked:    sr.GetIsRevoked(),
		ExpiresAt:    sr.GetExpiresAt().AsTime(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.SessionRes{
		Id:           sess.ID,
		UserEmail:    sess.UserEmail,
		RefreshToken: sess.RefreshToken,
		IsRevoked:    sess.IsRevoked,
		ExpiresAt:    timestamppb.New(sess.ExpiresAt),
	}, nil
}

func (s *Server) GetSession(ctx context.Context, sr *pb.SessionReq) (*pb.SessionRes, error) {
	sess, err := s.storer.GetSession(ctx, sr.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.SessionRes{
		Id:           sess.ID,
		UserEmail:    sess.UserEmail,
		RefreshToken: sess.RefreshToken,
		IsRevoked:    sess.IsRevoked,
		ExpiresAt:    timestamppb.New(sess.ExpiresAt),
	}, nil
}

func (s *Server) RevokeSession(ctx context.Context, sr *pb.SessionReq) (*pb.SessionRes, error) {
	err := s.storer.RevokeSession(ctx, sr.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.SessionRes{}, nil
}

func (s *Server) DeleteSession(ctx context.Context, sr *pb.SessionReq) (*pb.SessionRes, error) {
	err := s.storer.DeleteSession(ctx, sr.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.SessionRes{}, nil
}
