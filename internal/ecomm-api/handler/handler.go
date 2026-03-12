package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dockyguy56/ecomm/internal/ecomm-api/json"
	"github.com/dockyguy56/ecomm/internal/ecomm-grpc/pb"
	"github.com/dockyguy56/ecomm/internal/token"
	"github.com/dockyguy56/ecomm/internal/util"
	"github.com/go-chi/chi"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type handler struct {
	ctx        context.Context
	client     pb.EcommClient
	TokenMaker *token.JWTMaker
}

func NewHandler(ctx context.Context, client pb.EcommClient, secretKey string) *handler {
	return &handler{
		ctx:        ctx,
		client:     client,
		TokenMaker: token.NewJWTMaker(secretKey),
	}
}

func (h *handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p ProductRequest

	if err := json.Read(r, &p); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request payload: %w", err), http.StatusBadRequest)
		return
	}

	product, err := h.client.CreateProduct(h.ctx, toPBProductReq(p))
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

	product, err := h.client.GetProduct(h.ctx, &pb.ProductReq{Id: i})
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	response := toProductResponse(product)

	json.Write(w, http.StatusOK, response)
}

func (h *handler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.client.GetAllProducts(h.ctx, &pb.ProductReq{})
	if err != nil {
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	response := make([]ProductResponse, len(products.GetProducts()))
	for i, product := range products.GetProducts() {
		response[i] = toProductResponse(product)
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

	p.ID = i

	updated, err := h.client.UpdateProduct(h.ctx, toPBProductReq(p))
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

	_, err = h.client.DeleteProduct(h.ctx, &pb.ProductReq{Id: i})
	if err != nil {
		http.Error(w, fmt.Sprintf("Product not found: %s", err), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req OrderRequest

	if err := json.Read(r, &req); err != nil {
		http.Error(w, fmt.Errorf("Invalid request payload: %w", err).Error(), http.StatusBadRequest)
		return
	}

	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	order := toPBOrderReq(req)
	(*order).UserId = claims.ID

	created, err := h.client.CreateOrder(h.ctx, order)
	if err != nil {
		http.Error(w, fmt.Errorf("Failed to create order: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := toOrderResponse(created)

	json.Write(w, http.StatusCreated, response)
}

func (h *handler) GetAllOrdersByID(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	orders, err := h.client.GetAllOrders(h.ctx, &pb.OrderReq{UserId: claims.ID})
	if err != nil {
		http.Error(w, fmt.Sprintf("Order not found : %w", err), http.StatusNotFound)
		return
	}

	var ordersReponse ListOrderResponse
	for _, o := range orders.GetOrders() {
		ordersReponse.Orders = append(ordersReponse.Orders, toOrderResponse(o))
	}

	json.Write(w, http.StatusOK, ordersReponse)
}

func (h *handler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.client.GetAllOrders(h.ctx, &pb.OrderReq{})
	if err != nil {
		http.Error(w, fmt.Errorf("Failed to retrieve orders: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	var response []OrderResponse
	for _, order := range orders.GetOrders() {
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

	_, err = h.client.DeleteOrder(h.ctx, &pb.OrderReq{Id: i})
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

	created, err := h.client.CreateUser(h.ctx, toPBUserReq(u))
	if err != nil {
		http.Error(w, fmt.Errorf("error creating user : %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := toUserResponse(created)

	json.Write(w, http.StatusCreated, response)
}

func (h *handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.client.GetAllUsers(h.ctx, &pb.UserReq{})
	if err != nil {
		http.Error(w, fmt.Errorf("error getting all users: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	var response ListUserResponse
	for _, u := range users.GetUsers() {
		response.Users = append(response.Users, toUserResponse(u))
	}

	json.Write(w, http.StatusOK, response)
}

func (h *handler) updateUser(w http.ResponseWriter, r *http.Request) {
	var u UserRequest
	if err := json.Read(r, &u); err != nil {
		http.Error(w, fmt.Errorf("error decoding request body on update: %w", err).Error(), http.StatusBadRequest)
		return
	}

	claims := r.Context().Value(authKey{}).(*token.UserClaims)
	u.Email = claims.Email

	updatedUser, err := h.client.UpdateUser(h.ctx, toPBUserReq(u))
	if err != nil {
		http.Error(w, fmt.Errorf("error updating user: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := toUserResponse(updatedUser)
	json.Write(w, http.StatusOK, response)

}

func (h *handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, fmt.Errorf("error parsing ID: %w", err).Error(), http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteUser(h.ctx, &pb.UserReq{Id: i})
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

	gu, err := h.client.GetUser(h.ctx, &pb.UserReq{Email: u.Email})
	if err != nil {
		http.Error(w, fmt.Errorf("error geting user: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	err = util.CheckPassword(u.Password, gu.GetPassword())
	if err != nil {
		http.Error(w, fmt.Errorf("wrong password: %w", err).Error(), http.StatusUnauthorized)
		return
	}

	// create a json web token and return it as reponse
	accessToken, accessClaims, err := h.TokenMaker.CreateToken(gu.GetId(), gu.GetEmail(), gu.GetIsAdmin(), 15*time.Minute)
	if err != nil {
		http.Error(w, fmt.Errorf("error creating token: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, refreshClaims, err := h.TokenMaker.CreateToken(gu.GetId(), gu.GetEmail(), gu.GetIsAdmin(), 15*time.Minute)
	if err != nil {
		http.Error(w, fmt.Errorf("error creating token: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	session, err := h.client.CreateSession(h.ctx, &pb.SessionReq{
		Id:           refreshClaims.RegisteredClaims.ID,
		UserEmail:    gu.GetEmail(),
		RefreshToken: refreshToken,
		IsRevoked:    false,
		ExpiresAt:    timestamppb.New(refreshClaims.RegisteredClaims.ExpiresAt.Time),
	})
	if err != nil {
		http.Error(w, fmt.Errorf("error creating session: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	response := LoginUserResponse{
		SessionID:             session.GetId(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
		User:                  toUserResponse(gu),
	}

	json.Write(w, http.StatusOK, response)
}

func (h *handler) logoutUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	_, err := h.client.DeleteSession(h.ctx, &pb.SessionReq{Id: claims.RegisteredClaims.ID})
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

	refreshClaims, err := h.TokenMaker.VerifyToken(request.RefreshToken)
	if err != nil {
		http.Error(w, fmt.Errorf("error verifying token: %w", err).Error(), http.StatusUnauthorized)
		return
	}

	session, err := h.client.GetSession(h.ctx, &pb.SessionReq{Id: refreshClaims.RegisteredClaims.ID})
	if err != nil {
		http.Error(w, fmt.Errorf("error gettoing session: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	if session.GetIsRevoked() {
		http.Error(w, fmt.Errorf("session is revoked").Error(), http.StatusUnauthorized)
		return
	}

	if session.GetUserEmail() != refreshClaims.Email {
		http.Error(w, fmt.Errorf("invalid session").Error(), http.StatusUnauthorized)
		return
	}

	accessToken, accessClaims, err := h.TokenMaker.CreateToken(refreshClaims.ID, refreshClaims.Email, refreshClaims.IsAdmin, 24*time.Hour)
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
	claims := r.Context().Value(authKey{}).(*token.UserClaims)

	_, err := h.client.RevokeSession(h.ctx, &pb.SessionReq{Id: claims.RegisteredClaims.ID})
	if err != nil {
		http.Error(w, fmt.Errorf("error revoking session: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
