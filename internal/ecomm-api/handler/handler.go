package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dockyguy56/ecomm/internal/ecomm-api/server"
	"github.com/dockyguy56/ecomm/internal/ecomm-api/storer"
	"github.com/go-chi/chi"
)

type handler struct {
	ctx    context.Context
	server *server.Server
}

func NewHandler(ctx context.Context, server *server.Server) *handler {
	return &handler{ctx: ctx, server: server}
}

func (h *handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p ProductRequest

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	product, err := h.server.CreateProduct(h.ctx, toStorerProduct(p))
	if err != nil {
		http.Error(w, fmt.Errorf("failed to create product: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := toProductResponse(product)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.server.GetProductByID(h.ctx, i)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	response := toProductResponse(product)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.server.GetAllProducts(h.ctx)
	if err != nil {
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	response := make([]ProductResponse, len(products))
	for i, product := range products {
		response[i] = toProductResponse(&product)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var p ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	product, err := h.server.GetProductByID(h.ctx, i)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	patchProductRequest(product, p)
	updated, err := h.server.UpdateProduct(h.ctx, product)
	if err != nil {
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	response := toProductResponse(updated)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.server.DeleteProduct(h.ctx, i)
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toStorerProduct(p ProductRequest) *storer.Product {
	return &storer.Product{
		Name:         p.Name,
		Image:        p.Image,
		Category:     p.Category,
		Description:  p.Description,
		Rating:       p.Rating,
		NumReviews:   p.NumReviews,
		Price:        p.Price,
		CountInStock: p.CountInStock,
	}
}

func toProductResponse(p *storer.Product) ProductResponse {
	return ProductResponse{
		ID:           p.ID,
		Name:         p.Name,
		Image:        p.Image,
		Category:     p.Category,
		Description:  p.Description,
		Rating:       p.Rating,
		NumReviews:   p.NumReviews,
		Price:        p.Price,
		CountInStock: p.CountInStock,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

func patchProductRequest(p *storer.Product, req ProductRequest) {
	if req.Name != "" {
		p.Name = req.Name
	}

	if req.Image != "" {
		p.Image = req.Image
	}

	if req.Category != "" {
		p.Category = req.Category
	}

	if req.Description != "" {
		p.Description = req.Description
	}

	if req.Rating != 0 {
		p.Rating = req.Rating
	}

	if req.NumReviews != 0 {
		p.NumReviews = req.NumReviews
	}

	if req.Price != 0 {
		p.Price = req.Price
	}

	if req.CountInStock != 0 {
		p.CountInStock = req.CountInStock
	}
	p.UpdatedAt = toTimePtr(time.Now())
}

func toTimePtr(t time.Time) *time.Time {
	return &t
}

func (h *handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req OrderRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Errorf("Invalid request payload: %w", err).Error(), http.StatusBadRequest)
		return
	}

	order := toStorerOrder(req)
	created, err := h.server.CreateOrder(h.ctx, order)
	if err != nil {
		http.Error(w, fmt.Errorf("Failed to create order: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := toOrderResponse(created)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, fmt.Errorf("Invalid order ID: %w", err).Error(), http.StatusBadRequest)
		return
	}

	order, err := h.server.GetOrderByID(h.ctx, i)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	response := toOrderResponse(order)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.server.GetAllOrders(h.ctx)
	if err != nil {
		http.Error(w, fmt.Errorf("Failed to retrieve orders: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	var response []OrderResponse
	for _, order := range orders {
		response = append(response, toOrderResponse(order))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, fmt.Errorf("Invalid order ID: %w", err).Error(), http.StatusBadRequest)
		return
	}

	err = h.server.DeleteOrder(h.ctx, i)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toStorerOrder(req OrderRequest) *storer.Order {
	return &storer.Order{
		PaymentMethod: req.PaymentMethod,
		TaxPrice:      req.TaxPrice,
		ShippingPrice: req.ShippingPrice,
		TotalPrice:    req.TotalPrice,
		Items:         toStorerOrderItems(req.Items),
	}
}

func toStorerOrderItems(items []OrderItem) []storer.OrderItem {
	var res []storer.OrderItem
	for _, i := range items {
		res = append(res, storer.OrderItem{
			Name:      i.Name,
			Quantity:  i.Quantity,
			Image:     i.Image,
			Price:     i.Price,
			ProductID: i.ProductID,
		})
	}
	return res
}

func toOrderResponse(o *storer.Order) OrderResponse {
	return OrderResponse{
		ID:            o.ID,
		Items:         toOrderItemsResponse(o.Items),
		PaymentMethod: o.PaymentMethod,
		TaxPrice:      o.TaxPrice,
		ShippingPrice: o.ShippingPrice,
		TotalPrice:    o.TotalPrice,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}
}

func toOrderItemsResponse(items []storer.OrderItem) []OrderItem {
	var res []OrderItem
	for _, i := range items {
		res = append(res, OrderItem{
			Name:      i.Name,
			Quantity:  i.Quantity,
			Image:     i.Image,
			Price:     i.Price,
			ProductID: i.ProductID,
		})
	}
	return res
}
