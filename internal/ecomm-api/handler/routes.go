package handler

import (
	"net/http"

	"github.com/go-chi/chi"
)

var r *chi.Mux

func RegisterRoutes(h *handler) *chi.Mux {
	r = chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	r.Route("/products", func(r chi.Router) {
		r.Post("/", h.CreateProduct)
		r.Get("/", h.GetAllProducts)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetProductByID)
			r.Patch("/", h.UpdateProduct)
			r.Delete("/", h.DeleteProduct)
		})
	})

	r.Route("/orders", func(r chi.Router) {
		r.Post("/", h.CreateOrder)
		r.Get("/", h.GetAllOrders)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetOrderByID)
			r.Delete("/", h.DeleteOrder)
		})
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.CreateUser)
		r.Get("/", h.GetAllUsers)
		r.Patch("/", h.updateUser)

		r.Route("/{id}", func(r chi.Router) {
			r.Delete("/", h.DeleteUser)
		})

		r.Route("/login", func(r chi.Router) {
			r.Post("/", h.loginUser)
		})

		r.Route("/logout", func(r chi.Router) {
			r.Post("/", h.logoutUser)
		})

	})

	r.Route("/tokens", func(r chi.Router) {
		r.Route("/renew", func(r chi.Router) {
			r.Post("/", h.renewAccessToken)
		})

		r.Route("/revoke/{id}", func(r chi.Router) {
			r.Post("/", h.revokeSession)
		})
	})

	return r
}

func Start(addr string) error {
	return http.ListenAndServe(addr, r)
}
