package server

import (
	"context"

	"github.com/dockyguy56/ecomm/internal/ecomm-api/storer"
)

type Server struct {
	storer *storer.PostgresStorer
}

func NewServer(storer *storer.PostgresStorer) *Server {
	return &Server{storer: storer}
}

func (s *Server) CreateProduct(ctx context.Context, product *storer.Product) (*storer.Product, error) {
	return s.storer.CreateProduct(ctx, product)
}

func (s *Server) GetProductByID(ctx context.Context, id int64) (*storer.Product, error) {
	return s.storer.GetProductByID(ctx, id)
}

func (s *Server) GetAllProducts(ctx context.Context) ([]storer.Product, error) {
	return s.storer.GetAllProducts(ctx)
}

func (s *Server) UpdateProduct(ctx context.Context, product *storer.Product) (*storer.Product, error) {
	return s.storer.UpdateProduct(ctx, product)
}

func (s *Server) DeleteProduct(ctx context.Context, id int64) error {
	return s.storer.DeleteProduct(ctx, id)
}

func (s *Server) CreateOrder(ctx context.Context, order *storer.Order) (*storer.Order, error) {
	return s.storer.CreateOrder(ctx, order)
}

func (s *Server) GetOrderByID(ctx context.Context, id int64) (*storer.Order, error) {
	return s.storer.GetOrderByID(ctx, id)
}

func (s *Server) GetAllOrders(ctx context.Context) ([]*storer.Order, error) {
	return s.storer.GetAllOrders(ctx)
}

func (s *Server) DeleteOrder(ctx context.Context, id int64) error {
	return s.storer.DeleteOrder(ctx, id)
}
