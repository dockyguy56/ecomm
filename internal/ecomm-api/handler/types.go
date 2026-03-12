package handler

import "time"

type ProductRequest struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Image        string  `json:"image"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	Rating       int64   `json:"rating"`
	NumReviews   int64   `json:"num_reviews"`
	Price        float32 `json:"price"`
	CountInStock int64   `json:"count_in_stock"`
}

type ProductResponse struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	Image        string     `json:"image"`
	Category     string     `json:"category"`
	Description  string     `json:"description"`
	Rating       int64      `json:"rating"`
	NumReviews   int64      `json:"num_reviews"`
	Price        float32    `json:"price"`
	CountInStock int64      `json:"count_in_stock"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type OrderRequest struct {
	ID            int64        `json:"id"`
	Items         []*OrderItem `json:"items"`
	PaymentMethod string       `json:"payment_method"`
	TaxPrice      float32      `json:"tax_price"`
	ShippingPrice float32      `json:"shipping_price"`
	TotalPrice    float32      `json:"total_price"`
}

type OrderItem struct {
	Name      string  `json:"name"`
	Quantity  int64   `json:"quantity"`
	Image     string  `json:"image"`
	Price     float32 `json:"price"`
	ProductID int64   `json:"product_id"`
}

type OrderResponse struct {
	ID            int64        `json:"id"`
	Items         []*OrderItem `json:"items"`
	PaymentMethod string       `json:"payment_method"`
	TaxPrice      float32      `json:"tax_price"`
	ShippingPrice float32      `json:"shipping_price"`
	TotalPrice    float32      `json:"total_price"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     *time.Time   `json:"updated_at,omitempty"`
}

type ListOrderResponse struct {
	Orders []OrderResponse `json:"orders"`
}

type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type UserResponse struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

type ListUserResponse struct {
	Users []UserResponse `json:"users"`
}

type LoginUserRequest struct {
	Email    string `json:"email`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	SessionID             string       `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	RefreshToken          string       `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_toke_expires_at"`
	User                  UserResponse `json:"user"`
}

type RenewAccesTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RenewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}
