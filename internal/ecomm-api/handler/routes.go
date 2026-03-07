package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

var r *chi.Mux

func RegisterRoutes(h *handler) *chi.Mux {
	r = chi.NewRouter()
	tokenMaker := h.TokenMaker

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	r.Route("/products", func(r chi.Router) {
		r.With(GetAdminMiddlewareFnuc(tokenMaker)).Post("/", h.CreateProduct)
		r.Get("/", h.GetAllProducts)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetProductByID)
			r.Group(func(r chi.Router) {
				r.Use(GetAdminMiddlewareFnuc(tokenMaker))
				r.Patch("/", h.UpdateProduct)
				r.Delete("/", h.DeleteProduct)
			})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(GetAuthMiddlewareFnuc(tokenMaker))

		r.Get("/myorders", h.GetAllOrdersByID)
		r.Route("/orders", func(r chi.Router) {
			r.Post("/", h.CreateOrder)
			r.With(GetAdminMiddlewareFnuc(tokenMaker)).Get("/", h.GetAllOrders)

			r.Route("/{id}", func(r chi.Router) {
				r.Delete("/", h.DeleteOrder)
			})
		})
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.CreateUser)
		r.Post("/login", h.loginUser)

		r.Group(func(r chi.Router) {
			r.Use(GetAdminMiddlewareFnuc(tokenMaker))
			r.Get("/", h.GetAllUsers)

			r.Route("/{id}", func(r chi.Router) {
				r.Delete("/", h.DeleteUser)
			})

		})

		r.Group(func(r chi.Router) {
			r.Use(GetAuthMiddlewareFnuc(tokenMaker))
			r.Patch("/", h.updateUser)
			r.Post("/logout", h.logoutUser)
		})

	})

	r.Group(func(r chi.Router) {
		r.Use(GetAuthMiddlewareFnuc(tokenMaker))
		r.Route("/tokens", func(r chi.Router) {
			r.Post("/renew", h.renewAccessToken)
			r.Post("/revoke", h.revokeSession)
		})
	})

	return r
}

func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
