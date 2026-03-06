package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dockyguy56/ecomm/internal/ecomm-api/json"
	"github.com/dockyguy56/ecomm/internal/ecomm-api/server"
	"github.com/dockyguy56/ecomm/internal/ecomm-api/storer"
	"github.com/dockyguy56/ecomm/internal/token"
	"github.com/dockyguy56/ecomm/internal/util"
	"github.com/go-chi/chi"
)

type handler struct {
	ctx        context.Context
	server     *server.Server
	tokenMaker *token.JWTMaker
}

func NewHandler(ctx context.Context, server *server.Server, secretKey string) *handler {
	return &handler{
		ctx:        ctx,
		server:     server,
		tokenMaker: token.NewJWTMaker(secretKey),
	}
}

func (h *handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p ProductRequest

	if err := json.Read(r, &p); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	product, err := h.server.CreateProduct(h.ctx, toStorerProduct(p))
	if err != nil {
		http.Error(w, fmt.Errorf("failed to create product: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := toProductResponse(product)

	json.Write(w, http.StatusCreated, response)
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

	json.Write(w, http.StatusOK, response)
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

	json.Write(w, http.StatusOK, response)
}

func (h *handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var p ProductRequest
	if err := json.Read(r, &p); err != nil {
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

	json.Write(w, http.StatusOK, response)
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

	if err := json.Read(r, &req); err != nil {
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

	json.Write(w, http.StatusCreated, response)
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

	json.Write(w, http.StatusOK, response)
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

	json.Write(w, http.StatusOK, response)
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

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u UserRequest
	if err := json.Read(r, &u); err != nil {
		http.Error(w, fmt.Errorf("Bad request: %w", err).Error(), http.StatusBadRequest)
		return
	}

	// hash password
	hashed, err := util.HassPassword(u.Password)
	if err != nil {
		http.Error(w, fmt.Errorf("error hasing password: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	u.Password = hashed

	created, err := h.server.CreateUser(h.ctx, toStorerUser(u))
	if err != nil {
		http.Error(w, fmt.Errorf("error creating user : %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := toUserResponse(created)

	json.Write(w, http.StatusCreated, response)
}

func toStorerUser(u UserRequest) *storer.User {
	return &storer.User{
		Name:     u.Name,
		Email:    u.Email,
		Password: u.Password,
		IsAdmin:  u.IsAdmin,
	}
}

func toUserResponse(u *storer.User) UserResponse {
	return UserResponse{
		Name:    u.Name,
		Email:   u.Email,
		IsAdmin: u.IsAdmin,
	}
}

func (h *handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.server.GetAllUsers(h.ctx)
	if err != nil {
		http.Error(w, fmt.Errorf("error getting all users: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	var response ListUserRespose
	for _, u := range users {
		response.Users = append(response.Users, toUserResponse(&u))
	}

	json.Write(w, http.StatusOK, response)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var u UserRequest
	if err := json.Read(r, &u); err != nil {
		http.Error(w, fmt.Errorf("error decoding request body on update: %w", err).Error(), http.StatusBadRequest)
		return
	}

	user, err := h.server.GetUser(h.ctx, u.Email)
	if err != nil {
		http.Error(w, fmt.Errorf("error getting userL %w", err).Error(), http.StatusInternalServerError)
		return
	}

	patchUserRequest(user, u)

	updatedUser, err := h.server.UpdateUser(h.ctx, user)
	if err != nil {
		http.Error(w, fmt.Errorf("error updating user: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := toUserResponse(updatedUser)
	json.Write(w, http.StatusOK, response)

}
func patchUserRequest(user *storer.User, u UserRequest) {
	if u.Name != "" {
		user.Name = u.Name
	}
	if u.Email != "" {
		user.Email = u.Email
	}
	if u.Password != "" {
		hashed, err := util.HassPassword(u.Password)
		if err != nil {
			panic(err)
		}
		user.Password = hashed
	}
	if u.IsAdmin {
		user.IsAdmin = u.IsAdmin
	}
	user.UpdatedAt = toTimePtr(time.Now())
}

func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, fmt.Errorf("error parsing ID: %w", err).Error(), http.StatusBadRequest)
		return
	}

	err = h.server.DeleteUser(h.ctx, i)
	if err != nil {
		http.Error(w, fmt.Errorf("error deleting user: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) loginUser(w http.ResponseWriter, r *http.Request) {
	var u LoginUserRequest
	if err := json.Read(r, &u); err != nil {
		http.Error(w, fmt.Errorf("error decoding request body at LoginUser: %w", err).Error(), http.StatusBadRequest)
		return
	}

	gu, err := h.server.GetUser(h.ctx, u.Email)
	if err != nil {
		http.Error(w, fmt.Errorf("error geting user: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	err = util.CheckPassword(u.Password, gu.Password)
	if err != nil {
		http.Error(w, fmt.Errorf("wrong password: %w", err).Error(), http.StatusUnauthorized)
		return
	}

	// create a json web token and return it as reponse
	accessToken, accessClaims, err := h.tokenMaker.CreateToken(gu.ID, gu.Email, gu.IsAdmin, 15*time.Minute)
	if err != nil {
		http.Error(w, fmt.Errorf("error creating token: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, refreshClaims, err := h.tokenMaker.CreateToken(gu.ID, gu.Email, gu.IsAdmin, 15*time.Minute)
	if err != nil {
		http.Error(w, fmt.Errorf("error creating token: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	session, err := h.server.CreateSession(h.ctx, &storer.Session{
		ID:           refreshClaims.RegisteredClaims.ID,
		UserEmail:    gu.Email,
		RefreshToken: refreshToken,
		IsRevoked:    false,
		ExpiresAt:    refreshClaims.RegisteredClaims.ExpiresAt.Time,
	})
	if err != nil {
		http.Error(w, fmt.Errorf("error creating session: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := LoginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
		User:                  toUserResponse(gu),
	}

	json.Write(w, http.StatusOK, response)
}

func (h *handler) logoutUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, fmt.Errorf("missing session ID:").Error(), http.StatusBadRequest)
		return
	}

	err := h.server.DeleteSession(h.ctx, id)
	if err != nil {
		http.Error(w, fmt.Errorf("error deleting session: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) renewAccessToken(w http.ResponseWriter, r *http.Request) {
	var request RenewAccesTokenRequest
	if err := json.Read(r, &request); err != nil {
		http.Error(w, fmt.Errorf("error decoding request body at renew access token: %w", err).Error(), http.StatusBadRequest)
		return
	}

	refreshClaims, err := h.tokenMaker.VerifyToken(request.RefreshToken)
	if err != nil {
		http.Error(w, fmt.Errorf("error verifying token: %w", err).Error(), http.StatusUnauthorized)
		return
	}

	session, err := h.server.GetSession(h.ctx, refreshClaims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, fmt.Errorf("error gettoing session: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	if session.IsRevoked {
		http.Error(w, fmt.Errorf("session is revoked").Error(), http.StatusUnauthorized)
		return
	}

	if session.UserEmail != refreshClaims.Email {
		http.Error(w, fmt.Errorf("invalid session").Error(), http.StatusUnauthorized)
		return
	}

	accessToken, accessClaims, err := h.tokenMaker.CreateToken(refreshClaims.ID, refreshClaims.Email, refreshClaims.IsAdmin, 24*time.Hour)
	if err != nil {
		http.Error(w, fmt.Errorf("error creating token: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := RenewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessClaims.RegisteredClaims.ExpiresAt.Time,
	}

	json.Write(w, http.StatusOK, response)

}

func (h *handler) revokeSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, fmt.Errorf("missing session id").Error(), http.StatusBadRequest)
		return
	}

	err := h.server.RevokeSession(h.ctx, id)
	if err != nil {
		http.Error(w, fmt.Errorf("error revoking session: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
